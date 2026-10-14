package coinglass

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"
)

// 创建爬虫实例
var spider = NewSpider()

// TestCoinglassAPI 测试Coinglass API
func TestCoinglassAPI(t *testing.T) {

	// 测试币种市场数据API
	apiURL := "https://capi.coinglass.com/api/home/v2/coinMarkets?sort=&order=&keyword=&pageNum=1&pageSize=5&ex=all"

	fmt.Printf("🕷️ 测试 Coinglass API\n")
	fmt.Printf("URL: %s\n", apiURL)

	// 获取数据
	data, err := spider.GetData(apiURL)
	if err != nil {
		t.Errorf("获取数据失败: %v", err)
		return
	}

	preview := getDataPreview(data, 100)
	t.Logf("✅ 市场数据获取成功，数据长度: %d, 数据预览: %s", len(data), preview)
}

// 🔍 TestOpenInterestChart 测试持仓量图表数据接口
// 📊 获取指定币种的持仓量变化趋势数据
func TestOpenInterestChart(t *testing.T) {
	// 📋 构建查询参数
	params := map[string]string{
		"symbol":       "BTC", // 🪙 币种符号
		"timeType":     "0",   // ⏰ 时间类型
		"exchangeName": "",    // 🏢 交易所名称（空表示全部）
		"currency":     "USD", // 💵 计价货币
		"type":         "0",   // 📈 数据类型
	}

	// 🌐 构建完整的API URL
	apiURL := "https://capi.coinglass.com/api/openInterest/v3/chart"
	fullURL := buildURLWithParams(apiURL, params)

	// 🚀 发起请求获取数据
	result, err := spider.GetData(fullURL)
	if err != nil {
		t.Fatalf("❌ 获取持仓量图表数据失败: %v", err)
	}

	preview := getDataPreview(result, 100)
	t.Logf("✅ 持仓量图表数据获取成功，数据长度: %d, 数据预览: %s", len(result), preview)
}

// 📊 TestFuturesHomeStatistics 测试期货首页统计数据接口
// 🏠 获取期货市场的整体统计信息和概览数据
func TestFuturesHomeStatistics(t *testing.T) {
	// 🌐 构建API URL（无需额外参数）
	apiURL := "https://capi.coinglass.com/api/futures/home/statistics"

	// 🚀 发起请求获取数据
	result, err := spider.GetData(apiURL)
	if err != nil {
		t.Fatalf("❌ 获取期货首页统计数据失败: %v", err)
	}

	preview := getDataPreview(result, 100)
	t.Logf("✅ 期货首页统计数据获取成功，数据长度: %d, 数据预览: %s", len(result), preview)
}

// 🏢 TestDerivativeExchangeList 测试衍生品交易所列表接口
// 📋 获取支持衍生品交易的所有交易所信息
func TestDerivativeExchangeList(t *testing.T) {
	// 🌐 构建API URL（无需额外参数）
	apiURL := "https://capi.coinglass.com/api/derivative/exchange/list"

	// 🚀 发起请求获取数据
	result, err := spider.GetData(apiURL)
	if err != nil {
		t.Fatalf("❌ 获取衍生品交易所列表失败: %v", err)
	}

	preview := getDataPreview(result, 100)
	t.Logf("✅ 衍生品交易所列表获取成功，数据长度: %d, 数据预览: %s", len(result), preview)
}

// 💰 TestCoinMarkets 测试币种市场数据接口
// 📈 获取币种的市场排行和详细交易数据
func TestCoinMarkets(t *testing.T) {
	// 📋 构建查询参数
	params := map[string]string{
		"sort":     "",    // 🔄 排序字段（空表示默认排序）
		"order":    "",    // ⬆️ 排序方向（空表示默认方向）
		"keyword":  "",    // 🔍 搜索关键词（空表示不筛选）
		"pageNum":  "1",   // 📄 页码
		"pageSize": "20",  // 📊 每页数量
		"ex":       "all", // 🏢 交易所筛选（all表示全部）
	}

	// 🌐 构建完整的API URL
	apiURL := "https://capi.coinglass.com/api/home/v2/coinMarkets"
	fullURL := buildURLWithParams(apiURL, params)

	// 🚀 发起请求获取数据
	result, err := spider.GetData(fullURL)
	if err != nil {
		t.Fatalf("❌ 获取币种市场数据失败: %v", err)
	}

	preview := getDataPreview(result, 100)
	t.Logf("✅ 币种市场数据获取成功，数据长度: %d, 数据预览: %s", len(result), preview)
}

// 🔗 TestExchangeFuturesPairInfo 测试交易所期货交易对信息接口
// 📊 获取各大交易所的期货交易对详细配置信息
func TestExchangeFuturesPairInfo(t *testing.T) {
	// 🌐 构建API URL（无需额外参数）
	apiURL := "https://capi.coinglass.com/api/exchange/futures/pairInfo"

	// 🚀 发起请求获取数据
	result, err := spider.GetData(apiURL)
	if err != nil {
		t.Fatalf("❌ 获取交易所期货交易对信息失败: %v", err)
	}

	preview := getDataPreview(result, 100)
	t.Logf("✅ 交易所期货交易对信息获取成功，数据长度: %d, 数据预览: %s", len(result), preview)
}

// 🪙 TestSpotSupportCoin 测试现货支持币种接口
// 📋 获取平台支持的所有现货交易币种列表
func TestSpotSupportCoin(t *testing.T) {
	// 🌐 构建API URL（无需额外参数）
	apiURL := "https://capi.coinglass.com/api/spot/support/coin"

	// 🚀 发起请求获取数据
	result, err := spider.GetData(apiURL)
	if err != nil {
		t.Fatalf("❌ 获取现货支持币种失败: %v", err)
	}

	preview := getDataPreview(result, 100)
	t.Logf("✅ 现货支持币种获取成功，数据长度: %d, 数据预览: %s", len(result), preview)
}

// 🏦 TestETFFlow 测试比特币现货ETF净流入流出接口
// 📊 对应前台图表「比特币现货ETF净流入流出 (BTC)」
func TestETFFlow(t *testing.T) {
	// 🚀 获取全量历史数据
	days, err := spider.GetETFFlow()
	if err != nil {
		t.Fatalf("❌ 获取ETF净流入流出数据失败: %v", err)
	}

	if len(days) == 0 {
		t.Fatal("❌ 返回的ETF数据为空")
	}

	t.Logf("✅ ETF净流入流出数据获取成功，共 %d 天，区间 %s ~ %s",
		len(days),
		days[0].Time().Format("2006-01-02"),
		days[len(days)-1].Time().Format("2006-01-02"))

	// 📅 打印最近5天的总量
	fmt.Println("\n📅 最近5天比特币现货ETF净流入流出:")
	for _, d := range days[len(days)-5:] {
		fmt.Printf("  %s  %+12.2f BTC  (%+9.1f 百万美元)\n",
			d.Time().Format("2006-01-02"), d.Change, d.ChangeUsd/1e6)
	}

	// 🏢 打印最新一日的发行商明细
	latest := days[len(days)-1]
	fmt.Printf("\n🏢 %s 各发行商明细:\n", latest.Time().Format("2006-01-02"))
	for _, issuer := range latest.List {
		fmt.Printf("  %-6s %-16s %+12.2f BTC  %+9.1f 百万美元\n",
			issuer.Ticker, issuer.Name, issuer.Change, issuer.ChangeUsd/1e6)
	}
}

// 📉 TestETFSpotInFlow 测试比特币现货ETF净流入及资产快照接口
// 💡 精简版数据，无发行商拆分，但含ETF总净资产和BTC价格
func TestETFSpotInFlow(t *testing.T) {
	days, err := spider.GetETFSpotInFlow()
	if err != nil {
		t.Fatalf("❌ 获取ETF净流入数据失败: %v", err)
	}

	if len(days) == 0 {
		t.Fatal("❌ 返回的ETF数据为空")
	}

	latest := days[len(days)-1]
	t.Logf("✅ ETF净流入数据获取成功，共 %d 天", len(days))
	fmt.Printf("\n📊 %s 快照: 净流入 %+.1f 百万美元, 总净资产 %.1f 亿美元, BTC价格 $%.1f\n",
		latest.Time().Format("2006-01-02"),
		latest.Change/1e6, latest.NetAssets/1e8, latest.BtcPrice)
}

// 📋 TestETFList 测试比特币现货ETF清单接口
// 🏦 获取各ETF的发行商、托管方、AUM、管理费率等基本信息
func TestETFList(t *testing.T) {
	list, err := spider.GetETFList()
	if err != nil {
		t.Fatalf("❌ 获取ETF清单失败: %v", err)
	}

	if len(list) == 0 {
		t.Fatal("❌ 返回的ETF清单为空")
	}

	t.Logf("✅ ETF清单获取成功，共 %d 只", len(list))

	fmt.Println("\n🏦 比特币现货ETF清单（前5只）:")
	for i, etf := range list {
		if i >= 5 {
			break
		}
		fmt.Printf("  %-6s %-30s 发行商: %-12s AUM: %.1f 亿美元  费率: %.2f%%\n",
			etf.Ticker, strings.TrimSpace(etf.ETFName), etf.Issuer,
			etf.Aum/1e8, etf.Fee*100)
	}
}

// 🔥 TestCoinsHeatmap 测试币种热力图接口
// 📊 对应首页「热力图(24小时)」，验证三种指标类型
func TestCoinsHeatmap(t *testing.T) {
	types := []struct {
		key  string
		desc string
		unit string
	}{
		{HeatmapTypeVol, "24小时成交额", "亿美元"},
		{HeatmapTypeOI, "持仓量", "亿美元"},
		{HeatmapTypePrice, "24小时涨跌幅", "%"},
	}

	for _, ty := range types {
		coins, err := spider.GetCoinsHeatmap(ty.key)
		if err != nil {
			t.Fatalf("❌ 获取热力图失败 (type=%s): %v", ty.key, err)
		}

		if len(coins) == 0 {
			t.Fatalf("❌ 热力图数据为空 (type=%s)", ty.key)
		}

		t.Logf("✅ 热力图获取成功 (type=%s, %s)，共 %d 个币种", ty.key, ty.desc, len(coins))

		// 📊 按指标值降序，取前5名
		sorted := make([]HeatmapCoin, len(coins))
		copy(sorted, coins)
		sort.Slice(sorted, func(i, j int) bool { return sorted[i].Value > sorted[j].Value })

		fmt.Printf("\n🔥 %s 排行（前5）:\n", ty.desc)
		for _, c := range sorted[:5] {
			value := c.Value
			if ty.unit == "亿美元" {
				value /= 1e8
			}
			fmt.Printf("  %-8s %12.2f %-8s 24h涨跌 %+7.2f%%\n",
				c.Symbol, value, ty.unit, c.PriceChangePercent)
		}
	}
}

// 🚫 TestCoinsHeatmapInvalidType 测试热力图非法类型参数
// ⚠️ 服务端对非法type返回Server Error，客户端应提前拦截
func TestCoinsHeatmapInvalidType(t *testing.T) {
	if _, err := spider.GetCoinsHeatmap("marketCap"); err == nil {
		t.Fatal("❌ 非法类型应返回错误，但没有")
	} else {
		t.Logf("✅ 非法类型已被正确拦截: %v", err)
	}
}

// 😱 TestFearGreedIndex 测试恐惧贪婪指数接口
// 📊 对应前台「恐惧&贪婪指数」页，验证全量与最近N笔两种取法
func TestFearGreedIndex(t *testing.T) {
	all, err := spider.GetFearGreedIndex(0)
	if err != nil {
		t.Fatalf("❌ 获取恐惧贪婪指数失败: %v", err)
	}
	if len(all) == 0 {
		t.Fatal("❌ 恐惧贪婪指数为空")
	}

	t.Logf("✅ 恐惧贪婪指数全量获取成功，共 %d 笔，区间 %s ~ %s",
		len(all), all[0].Time().Format("2006-01-02"), all[len(all)-1].Time().Format("2006-01-02"))

	recent, err := spider.GetFearGreedIndex(7)
	if err != nil {
		t.Fatalf("❌ 获取最近7笔恐惧贪婪指数失败: %v", err)
	}
	if len(recent) != 7 {
		t.Fatalf("❌ 应返回7笔，实际 %d 笔", len(recent))
	}

	fmt.Println("\n😱 最近7天恐惧贪婪指数:")
	for _, p := range recent {
		fmt.Printf("  %s  %5.1f  %-6s  BTC $%.0f\n",
			p.Time().Format("2006-01-02"), p.Value, p.Level(), p.Price)
	}

	// 📊 全量各等级天数分布（同前端右侧统计）
	dist := map[string]int{}
	for _, p := range all {
		dist[p.Level()]++
	}
	fmt.Println("\n📊 历史等级分布:")
	for _, lv := range []string{"极度恐惧", "恐惧", "中性", "贪婪", "极度贪婪"} {
		fmt.Printf("  %-6s %5d 天 (%.2f%%)\n", lv, dist[lv], float64(dist[lv])/float64(len(all))*100)
	}
}

// 🧪 TestFearGreedIndexParseOffline 以内嵌样本离线验证恐惧贪婪指数解析（不发请求）
func TestFearGreedIndexParseOffline(t *testing.T) {
	raw := `[{"dates":[1517443200000,1517529600000,1517616000000],"values":[30,20.5,81],"prices":[10114.5,null,9000]},
		{"dates":[1],"values":[1],"prices":[null]}]`

	points, err := parseFearGreedIndex([]byte(raw))
	if err != nil {
		t.Fatalf("❌ 解析失败: %v", err)
	}
	if len(points) != 3 || points[0].Price != 10114.5 || points[1].Price != 0 || points[2].Value != 81 {
		t.Fatalf("❌ 解析结果不符: %+v", points)
	}
	if points[0].Time().Format("2006-01-02") != "2018-02-01" {
		t.Fatalf("❌ 时间转换错误: %s", points[0].Time())
	}

	// 🏷️ 等级分界：20 属极度恐惧、20.5 属恐惧、81 属极度贪婪
	for _, c := range []struct {
		v    float64
		want string
	}{{20, "极度恐惧"}, {20.5, "恐惧"}, {40, "恐惧"}, {60, "中性"}, {80, "贪婪"}, {81, "极度贪婪"}} {
		if got := (FearGreedPoint{Value: c.v}).Level(); got != c.want {
			t.Fatalf("❌ %.1f 应为 %s，实际 %s", c.v, c.want, got)
		}
	}

	if _, err := parseFearGreedIndex([]byte(`[{"dates":[1,2],"values":[1],"prices":[1,2]}]`)); err == nil {
		t.Fatal("❌ 数组长度不一致应返回错误，但没有")
	}
	t.Log("✅ 恐惧贪婪指数离线解析正常")
}

// 💵 TestGlobalM2SupplyGrowth 测试比特币 vs 全球M2货币供应增长接口
// 🌐 接口位于 fapi.coinglass.com，同时验证该域名的解密流程
func TestGlobalM2SupplyGrowth(t *testing.T) {
	points, err := spider.GetGlobalM2SupplyGrowth()
	if err != nil {
		t.Fatalf("❌ 获取全球M2供应增长失败: %v", err)
	}
	if len(points) == 0 {
		t.Fatal("❌ 全球M2供应增长数据为空")
	}

	first, latest := points[0], points[len(points)-1]
	if latest.M2Supply <= 0 || latest.Price <= 0 {
		t.Fatalf("❌ 最新一笔字段为空: %+v", latest)
	}

	t.Logf("✅ 全球M2供应增长获取成功，共 %d 周，区间 %s ~ %s",
		len(points), first.Time().Format("2006-01-02"), latest.Time().Format("2006-01-02"))

	fmt.Println("\n💵 最近5周全球M2:")
	for _, p := range points[len(points)-5:] {
		fmt.Printf("  %s  M2 %7.2f 万亿美元  同比 %+5.2f%%  BTC $%.0f\n",
			p.Time().Format("2006-01-02"), p.M2Supply/1e12, p.YoYGrowth, p.Price)
	}
}

// 💸 TestCoinFlow 测试BTC现货与合约资金流入流出
// 📊 对应前台 /spot-inflow-outflow 顶部图表（现货/合约两个切换）
func TestCoinFlow(t *testing.T) {
	markets := []struct {
		name string
		get  func(string, string, int) ([]FlowBar, error)
	}{
		{"现货", spider.GetSpotFlow},
		{"合约", spider.GetFuturesFlow},
	}

	for _, m := range markets {
		bars, err := m.get("BTC", "h24", 7)
		if err != nil {
			t.Fatalf("❌ 获取BTC%s流入流出失败: %v", m.name, err)
		}
		if len(bars) != 7 {
			t.Fatalf("❌ BTC%s应返回7笔，实际 %d 笔", m.name, len(bars))
		}

		t.Logf("✅ BTC%s流入流出获取成功，共 %d 笔", m.name, len(bars))
		fmt.Printf("\n💸 BTC%s最近7天流入流出（百万美元）:\n", m.name)
		for _, b := range bars {
			fmt.Printf("  %s  流入 %10.2f  流出 %10.2f  净流入 %+9.2f\n",
				b.Time().Format("2006-01-02"), b.InflowUSD/1e6, b.OutflowUSD/1e6, b.NetFlowUSD()/1e6)
		}
	}
}

// 🧪 TestCoinFlowOffline 离线检查资金流入流出的参数校验与解析（不发请求）
func TestCoinFlowOffline(t *testing.T) {
	if _, err := spider.GetSpotFlow("BTC", "d1", 10); err == nil {
		t.Fatal("❌ 非法时间粒度应返回错误，但没有")
	}
	if _, err := spider.GetFuturesFlow(" ", "h24", 10); err == nil {
		t.Fatal("❌ 空 symbol 应返回错误，但没有")
	}

	var bars []FlowBar
	if err := json.Unmarshal([]byte(`[[1791158400,1058428448.75,1171759252.41]]`), &bars); err != nil {
		t.Fatalf("❌ 解析失败: %v", err)
	}
	b := bars[0]
	if b.Time().Format("2006-01-02") != "2026-10-05" || b.InflowUSD != 1058428448.75 || b.NetFlowUSD() >= 0 {
		t.Fatalf("❌ 解析结果不符: %+v", b)
	}
	if err := json.Unmarshal([]byte(`[[1,2]]`), &bars); err == nil {
		t.Fatal("❌ 二元数组应返回错误，但没有")
	}
	t.Log("✅ 资金流入流出参数校验与解析正常")
}

// 💥 TestLiquidationMap 测试单一交易对清算地图（需要登录）
// 🔑 需设置环境变量 COINGLASS_OBE（浏览器 cookie「obe」的值），未设置则跳过，不发任何请求
// 🛡️ 只请求一次；遇到 ErrLoginRequired 直接失败、不重试，避免账号被风控
//
// 执行: COINGLASS_OBE=xxx go test -run TestLiquidationMap$ -v
func TestLiquidationMap(t *testing.T) {
	sp := newLoginSpider(t)

	m, err := sp.GetLiquidationMap("Binance_BTCUSDT", 1)
	if errors.Is(err, ErrLoginRequired) {
		t.Fatalf("🚪 obe 无效或已过期，请重新从浏览器取得: %v", err)
	}
	if err != nil {
		t.Fatalf("❌ 获取清算地图失败: %v", err)
	}
	if len(m.Levels) == 0 {
		t.Fatal("❌ 清算档位为空")
	}

	t.Logf("✅ 清算地图获取成功: %s %s，现价 %.0f，共 %d 个档位",
		m.Instrument.ExName, m.Instrument.InstrumentID, m.LastPrice, len(m.Levels))

	// 📊 按杠杆汇总，并列出金额最大的5个档位
	byLeverage := map[int]float64{}
	for _, l := range m.Levels {
		byLeverage[l.Leverage] += l.AmountUSD
	}
	fmt.Printf("\n⚡ 各杠杆清算总额:\n")
	for _, lev := range []int{10, 25, 50, 100} {
		fmt.Printf("  %3dx  %10.2f 百万美元\n", lev, byLeverage[lev]/1e6)
	}

	top := make([]LiqLevel, len(m.Levels))
	copy(top, m.Levels)
	sort.Slice(top, func(i, j int) bool { return top[i].AmountUSD > top[j].AmountUSD })
	fmt.Printf("\n💥 清算金额前5档位:\n")
	for _, l := range top[:min(5, len(top))] {
		fmt.Printf("  区间 %8.0f  价格 %8.0f  %8.2f 百万美元  %3dx  %s\n",
			l.Bucket, l.Price, l.AmountUSD/1e6, l.Leverage, l.Tag)
	}
}

// 💥 TestExchangeLiquidationMap 测试交易所聚合清算地图（需要登录）
// 🔑 同 TestLiquidationMap，需设置 COINGLASS_OBE
//
// 执行: COINGLASS_OBE=xxx go test -run TestExchangeLiquidationMap -v
func TestExchangeLiquidationMap(t *testing.T) {
	sp := newLoginSpider(t)

	m, err := sp.GetExchangeLiquidationMap("BTC", 1)
	if errors.Is(err, ErrLoginRequired) {
		t.Fatalf("🚪 obe 无效或已过期，请重新从浏览器取得: %v", err)
	}
	if err != nil {
		t.Fatalf("❌ 获取交易所清算地图失败: %v", err)
	}
	if len(m.Exchanges) == 0 {
		t.Fatal("❌ 交易所数据为空")
	}

	t.Logf("✅ 交易所清算地图获取成功: 现价 %.0f，范围 %.0f ~ %.0f，共 %d 家交易所",
		m.LastPrice, m.RangeLow, m.RangeHigh, len(m.Exchanges))

	// 📊 按交易所汇总，并区分现价下方（多单）与上方（空单）
	fmt.Printf("\n🏢 各交易所清算总额（百万美元）:\n")
	for _, ex := range m.Exchanges {
		var long, short float64
		for _, l := range ex.Levels {
			if l.Bucket < m.LastPrice {
				long += l.AmountUSD
			} else {
				short += l.AmountUSD
			}
		}
		fmt.Printf("  %-8s %-16s 档位 %3d  多单 %9.2f  空单 %9.2f\n",
			ex.Instrument.ExName, ex.Instrument.InstrumentID, len(ex.Levels), long/1e6, short/1e6)
	}
}

// 🚫 TestLiquidationMapOffline 离线检查参数校验与未登录拦截（不发请求）
func TestLiquidationMapOffline(t *testing.T) {
	sp := NewSpider()

	if _, err := sp.GetLiquidationMap("Binance_BTCUSDT", 1); !errors.Is(err, ErrLoginRequired) {
		t.Fatalf("❌ 未登录应直接返回 ErrLoginRequired，实际: %v", err)
	}

	if _, err := sp.GetExchangeLiquidationMap("BTC", 1); !errors.Is(err, ErrLoginRequired) {
		t.Fatalf("❌ 交易所版未登录应直接返回 ErrLoginRequired，实际: %v", err)
	}

	sp.SetLoginToken("dummy")
	if _, err := sp.GetLiquidationMap("Binance_BTCUSDT", 3); err == nil {
		t.Fatal("❌ 非法时间范围应返回错误，但没有")
	}
	if _, err := sp.GetExchangeLiquidationMap("", 1); err == nil {
		t.Fatal("❌ 空 symbol 应返回错误，但没有")
	}
	t.Log("✅ 参数校验与未登录拦截正常")
}

// 🧪 TestLiquidationMapParseOffline 以内嵌样本离线验证两种清算地图的解析（不发请求）
// 📝 样本结构取自 2026-10-05 实测响应，数值已简化
func TestLiquidationMapParseOffline(t *testing.T) {
	// 💥 单一交易对: liqMapV2 区间 -> [价格, 金额, 杠杆, 标签]，同区间多笔
	single := `{"instrument":{"exName":"Binance","instrumentId":"BTCUSDT"},"lastPrice":62000,
		"liqMapV2":{"62100":[[62080.5,2000000,100,"h1"],[62050,1000000,25,"h2"]],"61900":[[61910,500000,50,"h3"]]}}`

	m, err := parseLiquidationMap([]byte(single))
	if err != nil {
		t.Fatalf("❌ 解析单一交易对清算地图失败: %v", err)
	}
	if len(m.Levels) != 3 || m.LastPrice != 62000 || m.Instrument.InstrumentID != "BTCUSDT" {
		t.Fatalf("❌ 解析结果不符: %+v", m)
	}
	// 🪣 区间升序，同区间按杠杆升序
	if m.Levels[0].Bucket != 61900 || m.Levels[1].Leverage != 25 || m.Levels[2].Leverage != 100 {
		t.Fatalf("❌ 排序或 Bucket 回填错误: %+v", m.Levels)
	}
	if m.Levels[2].Price != 62080.5 || m.Levels[2].Tag != "h1" {
		t.Fatalf("❌ 档位字段解析错误: %+v", m.Levels[2])
	}

	// 🏢 交易所聚合: data 每家交易所一笔，杠杆与标签为 null
	ex := `{"instrument":{"exName":"Binance","instrumentId":"BTCUSDT"},"lastPrice":62000,"rangeHigh":70000,"rangeLow":55000,
		"data":[{"instrument":{"exName":"Binance","instrumentId":"BTCUSDT"},"liqMapV2":{"62100":[[62080,3000000,null,null]]}},
		{"instrument":{"exName":"OKX","instrumentId":"BTC-USDT-SWAP"},"liqMapV2":{"61000":[[61000,800000,null,null]],"63000":[[62990,900000,null,null]]}}]}`

	e, err := parseExchangeLiquidationMap([]byte(ex))
	if err != nil {
		t.Fatalf("❌ 解析交易所清算地图失败: %v", err)
	}
	if len(e.Exchanges) != 2 || e.RangeHigh != 70000 || e.RangeLow != 55000 {
		t.Fatalf("❌ 解析结果不符: %+v", e)
	}
	okx := e.Exchanges[1]
	if okx.Instrument.ExName != "OKX" || len(okx.Levels) != 2 || okx.Levels[0].Bucket != 61000 {
		t.Fatalf("❌ OKX 档位解析错误: %+v", okx)
	}
	if okx.Levels[0].Leverage != 0 || okx.Levels[0].Tag != "" {
		t.Fatalf("❌ null 杠杆/标签应为零值: %+v", okx.Levels[0])
	}

	// 🚫 非四元数组应报错
	if _, err := parseLiquidationMap([]byte(`{"liqMapV2":{"1":[[1,2,3]]}}`)); err == nil {
		t.Fatal("❌ 三元数组应返回错误，但没有")
	}
	t.Log("✅ 两种清算地图离线解析正常")
}

// 🔑 newLoginSpider 建立带登录凭证的爬虫，未设置 COINGLASS_OBE 时跳过测试
// 🎭 可选 COINGLASS_UA：填取得 obe 的浏览器 User-Agent，保持登录态一致
func newLoginSpider(t *testing.T) *Spider {
	t.Helper()

	token := os.Getenv("COINGLASS_OBE")
	if token == "" {
		t.Skip("⚠️ 未设置 COINGLASS_OBE，跳过需要登录的测试")
	}

	sp := NewSpider()
	sp.SetLoginToken(token)
	if ua := os.Getenv("COINGLASS_UA"); ua != "" {
		sp.SetUserAgent(ua)
	}
	return sp
}

// 🔧 buildURLWithParams 构建带参数的URL
// 📝 将参数映射转换为URL查询字符串并拼接到基础URL上
func buildURLWithParams(baseURL string, params map[string]string) string {
	if len(params) == 0 {
		return baseURL
	}

	url := baseURL + "?"
	first := true
	for key, value := range params {
		if !first {
			url += "&"
		}
		url += key + "=" + value
		first = false
	}
	return url
}

// 👀 getDataPreview 获取数据预览
// 📄 截取指定长度的数据内容用于预览，避免输出过长的数据
func getDataPreview(data string, maxLength int) string {
	if len(data) <= maxLength {
		return data
	}
	return data[:maxLength] + "..."
}
