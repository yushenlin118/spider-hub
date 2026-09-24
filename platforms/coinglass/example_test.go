package coinglass

import (
	"fmt"
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
