package coinglass

// 🌐 BaseURL Coinglass API基础URL
// 📝 所有接口均挂载在该域名下，与前台站点 www.coinglass.com 分离
const BaseURL = "https://capi.coinglass.com"

// 📊 通用数据接口端点
// 🔗 拼接方式: BaseURL + Endpoint + 查询参数
const (
	OpenInterestChartEndpoint   = "/api/openInterest/v3/chart"     // 📈 持仓量图表数据
	FuturesHomeStatsEndpoint    = "/api/futures/home/statistics"   // 🏠 期货首页统计
	DerivativeExchangeEndpoint  = "/api/derivative/exchange/list"  // 🏢 衍生品交易所列表
	CoinMarketsEndpoint         = "/api/home/v2/coinMarkets"       // 💰 币种市场数据
	ExchangeFuturesPairEndpoint = "/api/exchange/futures/pairInfo" // 🔗 期货交易对信息
	SpotSupportCoinEndpoint     = "/api/spot/support/coin"         // 🪙 现货支持币种
)

// 🏦 现货ETF接口端点
// 📝 注意命名规律: BTC接口不带币种前缀，其他币种才带（如 /api/etf/eth/flow）
const (
	// ETFFlowEndpoint 比特币现货ETF每日净流入流出（含各发行商明细）
	// 📊 对应前台图表「比特币现货ETF净流入流出 (BTC)」
	ETFFlowEndpoint = "/api/etf/flow"

	// ETFSpotInFlowEndpoint 比特币现货ETF每日净流入 + 总净资产 + 币价（精简版，无发行商拆分）
	ETFSpotInFlowEndpoint = "/api/stock/spot/inFlow"

	// ETFListEndpoint 比特币现货ETF清单（AUM、管理费、托管方等）
	ETFListEndpoint = "/api/bitcoin/etf/list"

	// ETFOverviewEndpoint 全币种ETF总览（流入排行榜）
	ETFOverviewEndpoint = "/api/etf/overview"

	// ETFHistoryChartEndpoint ETF持仓量/AUM历史图表
	// ⚠️ 这是持仓存量，不是每日净流入流出，勿与 ETFFlowEndpoint 混淆
	// 📝 必须带 type 参数，type=2 为有效值
	ETFHistoryChartEndpoint = "/api/etf/history/chart"

	// 其他币种的ETF每日净流入流出，结构与 ETFFlowEndpoint 一致
	ETFEthFlowEndpoint  = "/api/etf/eth/flow"  // 🔷 以太坊
	ETFSolFlowEndpoint  = "/api/etf/sol/flow"  // 🟣 Solana
	ETFXrpFlowEndpoint  = "/api/etf/xrp/flow"  // 💧 XRP
	ETFHypeFlowEndpoint = "/api/etf/hype/flow" // 🚀 HYPE
	ETFHkFlowEndpoint   = "/api/etf/hk/flow"   // 🇭🇰 香港ETF
)

// 🔥 热力图接口端点
const (
	// CoinsHeatmapEndpoint 币种热力图（首页「热力图(24小时)」数据源）
	// 📊 返回全部币种的24小时涨跌幅与指标值，用于绘制方块热力图
	// ⚠️ 必须带 type 参数，合法值见 HeatmapType* 常量；传数字或其他字符串会返回 Server Error
	CoinsHeatmapEndpoint = "/api/futures/coins/heatmap"

	// ExchangeHeatmapEndpoint 交易所热力图
	// 📝 需带 exName（如 Binance）与 type 参数
	// ⚠️ 实测 value 字段恒为0，方块大小指标疑似需要登录才下发
	ExchangeHeatmapEndpoint = "/api/derivative/exchange/heatmap"
)

// 🎨 HeatmapType 币种热力图的指标类型
// 📝 决定返回数据中 Value 字段的含义，PriceChangePercent 恒为24小时涨跌幅
const (
	HeatmapTypePrice = "price" // 📈 Value = 24小时涨跌幅（与 PriceChangePercent 相同）
	HeatmapTypeVol   = "vol"   // 💹 Value = 24小时成交额（美元）
	HeatmapTypeOI    = "oi"    // 🏦 Value = 持仓量（美元）
)

// ⏱️ 请求配置
const (
	RequestTimeoutSeconds = 30 // HTTP请求超时时间（秒）
)

// 🎭 RequestHeaders 模拟真实浏览器的请求头
// 🔐 其中 encryption 必须为 true，服务端才会返回加密数据并下发 user header 动态密钥
// ⚠️ cache-ts-v2 为毫秒时间戳，需在每次请求时动态生成，故不在此处定义
var RequestHeaders = map[string]string{
	"accept":             "application/json",
	"accept-language":    "zh-CN,zh;q=0.9",
	"encryption":         "true",
	"language":           "zh",
	"origin":             "https://www.coinglass.com",
	"priority":           "u=1, i",
	"referer":            "https://www.coinglass.com/",
	"sec-ch-ua":          `"Not)A;Brand";v="8", "Chromium";v="138", "Google Chrome";v="138"`,
	"sec-ch-ua-mobile":   "?0",
	"sec-ch-ua-platform": `"Windows"`,
	"sec-fetch-dest":     "empty",
	"sec-fetch-mode":     "cors",
	"sec-fetch-site":     "same-site",
	"user-agent":         "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36",
}

// 🔑 固定密钥来源（对应响应头 v 的特定取值）
// 📝 由前端 mn() 函数还原而来，服务端切换 v 时启用
const (
	KeySourceV55 = "170b070da9654622"
	KeySourceV66 = "d6537d845a964081"
	KeySourceV77 = "863f08689c97435b"
)
