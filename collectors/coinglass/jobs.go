package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/xieburoucoco/spider-hub/platforms/coinglass"
)

// ⏱️ Job 一个定期执行的抓取任务
type Job struct {
	Name       string        // 🏷️ 任务名称，同时写入 coinglass_collector_run.job
	Every      time.Duration // 🔁 成功后多久再跑
	RetryAfter time.Duration // 🔄 失败后多久重试（登录失效除外，见 runDue）
	NeedLogin  bool          // 🔒 需要 COINGLASS_OBE，未设置时跳过
	Run        func(ctx context.Context, c *Collector) (int, error)
}

// 🧰 Collector 任务执行时共用的依赖
// 🛡️ 公开任务与登录任务用不同的 Spider：登录 token 只给真正需要的清算地图，
// 避免每小时的公开请求也带着账号（增加风控曝光），也避免公开请求被登录限速拖慢
type Collector struct {
	spider      *coinglass.Spider // 🔓 匿名，供不需要登录的任务使用
	loginSpider *coinglass.Spider // 🔒 带 obe，仅 NeedLogin 任务使用；未设置 token 时为 nil
	store       *Store
}

// 💸 流入流出要收集的序列
var (
	flowSymbols   = []string{"BTC"}
	flowIntervals = map[string]time.Duration{"h1": time.Hour, "h24": 24 * time.Hour}
)

// 📥 flowBackfillLimits 首次回补依序尝试的笔数
// ⚠️ 实测合约 h24 超过约 3000 笔服务端直接报错（无 user header），现货 h24 / h1 则可到 4320
var flowBackfillLimits = []int{coinglass.FlowMaxLimit, 3000, 2000, 1000}

const (
	flowIncrementalLimit = 48 // 📏 增量抓取的最少笔数（h1 即 2 天），留重叠避免漏抓
	fearGreedIncremental = 30 // 📏 恐惧贪婪指数增量笔数
)

// 💥 清算地图要收集的快照
var liqTargets = struct {
	PairSymbol     string
	ExchangeSymbol string
	RangeDays      int
}{"Binance_BTCUSDT", "BTC", 1}

// 📋 Jobs 全部任务
var Jobs = []Job{
	{Name: "etf_flow", Every: 6 * time.Hour, RetryAfter: 15 * time.Minute, Run: runETFFlow},
	{Name: "etf_spot_inflow", Every: 6 * time.Hour, RetryAfter: 15 * time.Minute, Run: runETFSpotInFlow},
	{Name: "fear_greed", Every: 6 * time.Hour, RetryAfter: 15 * time.Minute, Run: runFearGreed},
	{Name: "global_m2", Every: 24 * time.Hour, RetryAfter: time.Hour, Run: runGlobalM2},
	{Name: "coin_flow", Every: time.Hour, RetryAfter: 10 * time.Minute, Run: runCoinFlow},
	{Name: "liq_map", Every: 24 * time.Hour, RetryAfter: time.Hour, NeedLogin: true, Run: runLiqMap},
}

// 📅 utcDate 毫秒时间戳 -> UTC 日期（time.Time 时分秒归零）
func utcDate(ms int64) time.Time {
	t := time.UnixMilli(ms).UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// 🏦 runETFFlow 比特币现货ETF每日流入流出（全量 upsert，约 600KB）
func runETFFlow(ctx context.Context, c *Collector) (int, error) {
	days, err := c.spider.GetETFFlow()
	if err != nil {
		return 0, err
	}

	b := &pgx.Batch{}
	for _, d := range days {
		date := utcDate(d.Date)
		b.Queue(`INSERT INTO public.coinglass_etf_flow_daily (date, net_flow_btc, net_flow_usd)
			VALUES ($1, $2, $3)
			ON CONFLICT (date) DO UPDATE SET net_flow_btc = EXCLUDED.net_flow_btc,
				net_flow_usd = EXCLUDED.net_flow_usd, updated_at = now()`,
			date, d.Change, d.ChangeUsd)
		for _, is := range d.List {
			b.Queue(`INSERT INTO public.coinglass_etf_flow_issuer_daily (date, ticker, issuer, net_flow_btc, net_flow_usd)
				VALUES ($1, $2, $3, $4, $5)
				ON CONFLICT (date, ticker) DO UPDATE SET issuer = EXCLUDED.issuer,
					net_flow_btc = EXCLUDED.net_flow_btc, net_flow_usd = EXCLUDED.net_flow_usd, updated_at = now()`,
				date, is.Ticker, is.Name, is.Change, is.ChangeUsd)
		}
	}
	return c.store.execBatch(ctx, b)
}

// 📉 runETFSpotInFlow 比特币现货ETF每日净流入 + 总净资产（全量 upsert）
func runETFSpotInFlow(ctx context.Context, c *Collector) (int, error) {
	days, err := c.spider.GetETFSpotInFlow()
	if err != nil {
		return 0, err
	}

	b := &pgx.Batch{}
	for _, d := range days {
		b.Queue(`INSERT INTO public.coinglass_etf_spot_inflow_daily (date, net_flow_usd, net_assets_usd, btc_price_usd)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (date) DO UPDATE SET net_flow_usd = EXCLUDED.net_flow_usd,
				net_assets_usd = EXCLUDED.net_assets_usd, btc_price_usd = EXCLUDED.btc_price_usd, updated_at = now()`,
			utcDate(d.Date), d.Change, d.NetAssets, d.BtcPrice)
	}
	return c.store.execBatch(ctx, b)
}

// 😱 runFearGreed 恐惧贪婪指数（首次全量，之后取最近 30 笔）
func runFearGreed(ctx context.Context, c *Collector) (int, error) {
	has, err := c.store.HasRows(ctx, "fear_greed_daily")
	if err != nil {
		return 0, err
	}
	limit := fearGreedIncremental
	if !has {
		limit = 0 // 📥 全量回补
	}

	points, err := c.spider.GetFearGreedIndex(limit)
	if err != nil {
		return 0, err
	}

	b := &pgx.Batch{}
	for _, p := range points {
		// 🪙 解析时 null 价格会变成 0，入库还原为 NULL
		var price *float64
		if p.Price != 0 {
			price = &p.Price
		}
		b.Queue(`INSERT INTO public.coinglass_fear_greed_daily (date, value, level, btc_price_usd)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (date) DO UPDATE SET value = EXCLUDED.value, level = EXCLUDED.level,
				btc_price_usd = EXCLUDED.btc_price_usd, updated_at = now()`,
			utcDate(p.Date), p.Value, p.Level(), price)
	}
	return c.store.execBatch(ctx, b)
}

// 💵 runGlobalM2 全球M2（接口只给全量，约 700 笔，直接 upsert）
func runGlobalM2(ctx context.Context, c *Collector) (int, error) {
	points, err := c.spider.GetGlobalM2SupplyGrowth()
	if err != nil {
		return 0, err
	}

	b := &pgx.Batch{}
	for _, p := range points {
		b.Queue(`INSERT INTO public.coinglass_global_m2_weekly (week, m2_supply_usd, yoy_growth_pct, btc_price_usd)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (week) DO UPDATE SET m2_supply_usd = EXCLUDED.m2_supply_usd,
				yoy_growth_pct = EXCLUDED.yoy_growth_pct, btc_price_usd = EXCLUDED.btc_price_usd, updated_at = now()`,
			utcDate(p.Timestamp), p.M2Supply, p.YoYGrowth, p.Price)
	}
	return c.store.execBatch(ctx, b)
}

// 💸 runCoinFlow 现货与合约流入流出（首次回补到服务端上限，之后增量）
func runCoinFlow(ctx context.Context, c *Collector) (int, error) {
	markets := []struct {
		name string
		get  func(string, string, int) ([]coinglass.FlowBar, error)
	}{
		{"spot", c.spider.GetSpotFlow},
		{"futures", c.spider.GetFuturesFlow},
	}

	total := 0
	now := time.Now()
	for _, m := range markets {
		for _, sym := range flowSymbols {
			for iv, dur := range flowIntervals {
				latest, err := c.store.LatestCoinFlow(ctx, m.name, sym, iv)
				if err != nil {
					return total, err
				}
				limits := flowBackfillLimits // 📥 无资料：回补服务端允许的最大历史
				if latest != nil {
					limits = []int{incrementalFlowLimit(*latest, dur, now)}
				}

				var bars []coinglass.FlowBar
				for _, limit := range limits {
					if bars, err = m.get(sym, iv, limit); err == nil {
						break
					}
				}
				if err != nil {
					return total, fmt.Errorf("%s %s %s: %w", m.name, sym, iv, err)
				}
				bars = trimLeadingEmpty(bars)

				b := &pgx.Batch{}
				for _, bar := range bars {
					final := !bar.Time().Add(dur).After(now)
					b.Queue(`INSERT INTO public.coinglass_coin_flow (market, symbol, interval, bar_time, inflow_usd, outflow_usd, is_final)
						VALUES ($1, $2, $3, $4, $5, $6, $7)
						ON CONFLICT (market, symbol, interval, bar_time) DO UPDATE SET inflow_usd = EXCLUDED.inflow_usd,
							outflow_usd = EXCLUDED.outflow_usd, is_final = EXCLUDED.is_final, updated_at = now()`,
						m.name, sym, iv, bar.Time(), bar.InflowUSD, bar.OutflowUSD, final)
				}
				n, err := c.store.execBatch(ctx, b)
				total += n
				if err != nil {
					return total, err
				}
			}
		}
	}
	return total, nil
}

// 📏 incrementalFlowLimit 依最新一根K线距今多久，算出这次要抓几根
// 🔄 收集器停机多久都能自动补齐（最多到服务端上限，h1 约 180 天），不足 flowIncrementalLimit 时取最少值
// 📝 回补时最新那根K线会被重抓一次（它当时可能还没收盘），所以 +1 再加少量余量
func incrementalFlowLimit(latest time.Time, dur time.Duration, now time.Time) int {
	n := int(now.Sub(latest)/dur) + 1 + 3
	return min(max(n, flowIncrementalLimit), coinglass.FlowMaxLimit)
}

// ✂️ trimLeadingEmpty 去掉开头流入流出皆为 0 的K线
// 📝 请求范围早于该市场有数据的时间时，服务端以 0 补齐，并非真实成交为 0
func trimLeadingEmpty(bars []coinglass.FlowBar) []coinglass.FlowBar {
	for i, b := range bars {
		if b.InflowUSD != 0 || b.OutflowUSD != 0 {
			return bars[i:]
		}
	}
	return nil
}

// 💥 liqSnapshot 待写入的清算地图快照
type liqSnapshot struct {
	MapType     string
	QuerySymbol string
	RangeDays   int
	LastPrice   float64
	RangeLow    *float64
	RangeHigh   *float64
	Levels      []liqLevelRow
}

type liqLevelRow struct {
	Exchange     string
	InstrumentID string
	Bucket       float64
	Price        float64
	AmountUSD    float64
	Leverage     *int
	Tag          *string
}

// 💥 runLiqMap 清算地图快照（交易对版 + 交易所聚合版）
// 🛡️ Spider 登录状态下自动限速，两次请求间隔至少 LoginMinRequestIntervalSeconds
func runLiqMap(ctx context.Context, c *Collector) (int, error) {
	t := liqTargets
	total := 0

	pair, err := c.loginSpider.GetLiquidationMap(t.PairSymbol, t.RangeDays)
	if err != nil {
		return 0, fmt.Errorf("pair %s: %w", t.PairSymbol, err)
	}
	snap := liqSnapshot{MapType: "pair", QuerySymbol: t.PairSymbol, RangeDays: t.RangeDays, LastPrice: pair.LastPrice}
	for _, l := range pair.Levels {
		lev, tag := l.Leverage, l.Tag
		snap.Levels = append(snap.Levels, liqLevelRow{pair.Instrument.ExName, pair.Instrument.InstrumentID,
			l.Bucket, l.Price, l.AmountUSD, &lev, &tag})
	}
	n, err := c.store.insertLiqSnapshot(ctx, snap)
	total += n
	if err != nil {
		return total, err
	}

	ex, err := c.loginSpider.GetExchangeLiquidationMap(t.ExchangeSymbol, t.RangeDays)
	if err != nil {
		return total, fmt.Errorf("exchange %s: %w", t.ExchangeSymbol, err)
	}
	snap = liqSnapshot{MapType: "exchange", QuerySymbol: t.ExchangeSymbol, RangeDays: t.RangeDays,
		LastPrice: ex.LastPrice, RangeLow: &ex.RangeLow, RangeHigh: &ex.RangeHigh}
	for _, e := range ex.Exchanges {
		for _, l := range e.Levels {
			// ⚠️ 聚合版的杠杆与标签服务端为 null，入库保持 NULL
			snap.Levels = append(snap.Levels, liqLevelRow{e.Instrument.ExName, e.Instrument.InstrumentID,
				l.Bucket, l.Price, l.AmountUSD, nil, nil})
		}
	}
	n, err = c.store.insertLiqSnapshot(ctx, snap)
	return total + n, err
}

// 🚪 isLoginError 是否为登录失效
func isLoginError(err error) bool {
	return errors.Is(err, coinglass.ErrLoginRequired)
}
