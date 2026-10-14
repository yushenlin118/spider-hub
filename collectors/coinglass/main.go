// 🕷️ Coinglass 收集器：定期抓取 platforms/coinglass 的数据写入 PostgreSQL
//
// ⚠️ 这是 spider-hub 里唯一可部署的程序，其余 platforms/ 仍是范例
//
// 环境变数:
//
//	DATABASE_URL   - 🗄️ 必填，如 postgres://user:pass@host:5432/db
//	COINGLASS_OBE  - 🔑 选填，浏览器 cookie「obe」；未设置时跳过清算地图
//	COINGLASS_UA   - 🎭 选填，取得 obe 的浏览器 User-Agent（只用于登录请求）
//
// 用法:
//
//	go run ./collectors/coinglass                 # 常驻，依排程执行
//	go run ./collectors/coinglass -once           # 全部任务立即跑一次后结束
//	go run ./collectors/coinglass -once -job liq_map
//	go run ./collectors/coinglass -migrate         # 只建表与更新注解
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/xieburoucoco/spider-hub/platforms/coinglass"
)

// ⏱️ checkInterval 排程检查间隔
const checkInterval = time.Minute

func main() {
	once := flag.Bool("once", false, "所有任务立即执行一次后结束（不看排程）")
	migrate := flag.Bool("migrate", false, "只套用 schema.sql（建表与注解）后结束，不抓数据")
	only := flag.String("job", "", "只执行指定任务，可用逗号分隔: "+jobNames())
	flag.Parse()

	if err := run(*once, *migrate, *only); err != nil {
		log.Fatal(err)
	}
}

func run(once, migrate bool, only string) error {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return fmt.Errorf("❌ 未设置 DATABASE_URL")
	}

	jobs, err := selectJobs(only)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	store, err := NewStore(ctx, dbURL)
	if err != nil {
		return err
	}
	defer store.Close()
	log.Println("✅ 资料库已连线，资料表就绪")
	if migrate {
		return nil
	}

	c := &Collector{spider: coinglass.NewSpider(), store: store}
	hasLogin := false
	if token := os.Getenv("COINGLASS_OBE"); token != "" {
		c.loginSpider = coinglass.NewSpider()
		c.loginSpider.SetLoginToken(token)
		if ua := os.Getenv("COINGLASS_UA"); ua != "" {
			c.loginSpider.SetUserAgent(ua)
		}
		hasLogin = true
	} else {
		log.Println("⚠️ 未设置 COINGLASS_OBE，需要登录的任务（liq_map）将跳过")
	}

	if once {
		for _, j := range jobs {
			if ctx.Err() != nil {
				break
			}
			execute(ctx, c, j, hasLogin)
		}
		return nil
	}

	log.Printf("🚀 常驻模式启动，每 %s 检查一次排程", checkInterval)
	for {
		runDue(ctx, c, jobs, hasLogin)
		select {
		case <-ctx.Done():
			log.Println("👋 收到结束信号，退出")
			return nil
		case <-time.After(checkInterval):
		}
	}
}

// 🔁 runDue 依序执行到期的任务
// 📝 是否到期依 coinglass_collector_run 最近一笔纪录判断，因此重启不会重复抓
//
//	成功       -> 距上次 >= Every 才再跑
//	登录失效   -> 同样等 Every（token 换新前重试只会增加封号风险）
//	其他失败   -> 距上次 >= RetryAfter 就重试
func runDue(ctx context.Context, c *Collector, jobs []Job, hasLogin bool) {
	for _, j := range jobs {
		if ctx.Err() != nil {
			return
		}
		if j.NeedLogin && !hasLogin {
			continue
		}

		last, err := c.store.LastRun(ctx, j.Name)
		if err != nil {
			log.Printf("❌ [%s] %v", j.Name, err)
			continue
		}
		if last != nil {
			wait := j.Every
			if last.Status == "error" && !last.LoginError {
				wait = j.RetryAfter
			}
			if time.Since(last.StartedAt) < wait {
				continue
			}
		}
		execute(ctx, c, j, hasLogin)
	}
}

// ▶️ execute 执行单一任务并写入执行纪录
func execute(ctx context.Context, c *Collector, j Job, hasLogin bool) {
	if j.NeedLogin && !hasLogin {
		log.Printf("⏭️ [%s] 需要登录，跳过", j.Name)
		return
	}

	started := time.Now()
	rows, err := j.Run(ctx, c)
	loginErr := err != nil && isLoginError(err)

	switch {
	case loginErr:
		log.Printf("🚪 [%s] 登录失效，请更新 COINGLASS_OBE: %v", j.Name, err)
	case err != nil:
		log.Printf("❌ [%s] 失败（%s 后重试）: %v", j.Name, j.RetryAfter, err)
	default:
		log.Printf("✅ [%s] 完成，写入 %d 笔，耗时 %s", j.Name, rows, time.Since(started).Round(time.Millisecond))
	}

	// 📋 结束信号下也要记下纪录，用独立 context
	recCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if recErr := c.store.RecordRun(recCtx, j.Name, started, rows, err, loginErr); recErr != nil {
		log.Printf("❌ [%s] %v", j.Name, recErr)
	}
}

// 🎯 selectJobs 依 -job 参数筛选任务
func selectJobs(only string) ([]Job, error) {
	if only == "" {
		return Jobs, nil
	}
	var picked []Job
	for _, name := range strings.Split(only, ",") {
		name = strings.TrimSpace(name)
		found := false
		for _, j := range Jobs {
			if j.Name == name {
				picked = append(picked, j)
				found = true
			}
		}
		if !found {
			return nil, fmt.Errorf("❌ 未知任务 %q（可选: %s）", name, jobNames())
		}
	}
	return picked, nil
}

func jobNames() string {
	names := make([]string, len(Jobs))
	for i, j := range Jobs {
		names[i] = j.Name
	}
	return strings.Join(names, ", ")
}
