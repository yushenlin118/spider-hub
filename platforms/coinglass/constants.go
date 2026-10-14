package coinglass

// 🌐 BaseURL Coinglass API基础URL
// 📝 所有接口均挂载在该域名下，与前台站点 www.coinglass.com 分离
const BaseURL = "https://capi.coinglass.com"

// 🌐 FapiBaseURL 部分宏观指标接口所在的另一个域名
// 📝 前端直接写死完整URL（如 M2 供应增长），加密协议与 BaseURL 完全相同，GetData 可直接请求
const FapiBaseURL = "https://fapi.coinglass.com"

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

// 😱 恐惧贪婪指数接口端点（前台 /pro/i/FearGreedIndex）
const (
	// FearGreedIndexEndpoint 恐惧贪婪指数历史
	// 📝 参数 size：空字符串为全量（实测自 2018-02-01 起约每日一笔），数字为最近 N 笔
	// 📦 返回数组，前端只用 data[0]（dates/values/prices 三个等长数组）
	// ⚠️ data[1]、data[2] 是另外两条序列，前端未使用、含义未知，不要自行解读
	// 🔓 不需要登录、不需要签名
	FearGreedIndexEndpoint = "/api/index/history"
)

// 💵 宏观指标接口端点（挂在 FapiBaseURL 下）
const (
	// GlobalM2SupplyGrowthEndpoint 比特币 vs 全球M2货币供应增长（前台 /pro/i/bitcoin-m2-supply-growth）
	// 📝 不带参数，单次返回全量；每周一笔，实测自 2013-05-20 起
	// ⚠️ M2 为宏观数据，最新一笔通常落后当前日期约一个月
	// 🔓 不需要登录
	GlobalM2SupplyGrowthEndpoint = "/api/index/globalM2SupplyGrowth"
)

// 💸 资金流入流出接口端点（前台 /spot-inflow-outflow 顶部图表）
const (
	// FlowKlineEndpoint 主动买卖成交额K线，前端用来画「现货/合约流入流出」柱状图
	// 📝 参数: symbol（格式见 FlowSymbolFormat）、interval、limit、diff=false、minLimit=false
	// 📦 返回 [[秒级时间戳, 流入USD, 流出USD], ...]，按时间升序；前端净流入 = 流入 - 流出
	// 🔓 不需要登录
	FlowKlineEndpoint = "/api/v2/kline"

	// FlowSymbolFormat symbol 参数格式: ALL#<币种>#<市场>
	FlowSymbolFormat = "ALL#%s#%s"

	FlowMarketSpot    = "aggregated_spot_buy_sell_usd" // 🏪 现货（页面切到「现货」）
	FlowMarketFutures = "aggregated_buy_sell_usd"      // 📜 合约（页面切到「合约」）
)

// 📏 资金流入流出分页参数
const (
	FlowDefaultLimit = 300  // 📄 默认笔数（与前端一致）
	FlowMaxLimit     = 4320 // 🔝 实测服务端单次最多返回 4320 笔，超过会被截断
)

// ⏱️ FlowIntervals 资金流入流出可用的时间粒度（还原自前端时间选项）
// 📝 实测 d1 等未列出的值服务端也接受，但前端没用，这里不放行
var FlowIntervals = map[string]bool{
	"m1": true, "m5": true, "m15": true, "m30": true,
	"h1": true, "h2": true, "h4": true, "h6": true, "h8": true, "h12": true, "h24": true,
	"w1": true,
}

// 🌡️ 恐惧贪婪等级分界（还原自前端 onChangeOption，区间为左开右闭，如 (20, 40] 为恐惧）
const (
	FearGreedExtremeFearMax = 20 // 😱 <= 20 极度恐惧
	FearGreedFearMax        = 40 // 😨 <= 40 恐惧
	FearGreedNeutralMax     = 60 // 😐 <= 60 中性
	FearGreedGreedMax       = 80 // 😀 <= 80 贪婪，以上为极度贪婪
)

// 🎨 HeatmapType 币种热力图的指标类型
// 📝 决定返回数据中 Value 字段的含义，PriceChangePercent 恒为24小时涨跌幅
const (
	HeatmapTypePrice = "price" // 📈 Value = 24小时涨跌幅（与 PriceChangePercent 相同）
	HeatmapTypeVol   = "vol"   // 💹 Value = 24小时成交额（美元）
	HeatmapTypeOI    = "oi"    // 🏦 Value = 持仓量（美元）
)

// 💥 清算地图接口端点（前台 /pro/futures/LiquidationMap）
// 🔒 两者均需登录：未带有效 obe 时服务端返回 code=40000（前端据此弹出登录框）
const (
	// LiqMapEndpoint 单一交易对清算地图
	// 📝 参数: merge=true、symbol（Binance_BTCUSDT 格式）、interval、limit、data（TOTP签名，见 generateLiqSign）
	LiqMapEndpoint = "/api/index/5/liqMap"

	// ExLiqMapEndpoint 交易所聚合清算地图
	// 📝 参数: merge=true、symbol（币种，如 BTC）、interval、limit；不需要 data 签名
	ExLiqMapEndpoint = "/api/index/2/exLiqMap"
)

// 🔐 清算接口 data 签名参数（还原自前端，与清算热力图共用）
// 📐 流程: TOTP(secret, step=30) -> "<unix秒>,<6位码>" -> AES-256-ECB + PKCS7 -> base64
const (
	LiqSignTOTPSecret = "I65VU7K5ZQL7WB4E"                 // 🔑 TOTP 密钥（base32）
	LiqSignTOTPStep   = 30                                 // ⏱️ TOTP 时间步长（秒）
	LiqSignAESKey     = "1f68efd73f8d4921acc0dead41dd39bc" // 🔒 AES-256 密钥（32字节字符串直接当key）
)

// 📅 LiqMapRanges 清算地图时间范围 -> 请求参数 interval/limit 的对照
// 📝 还原自前端 onChangeTime()，key 为页面上的天数选项
// ⚠️ 7天对应 interval=5 不是 7，90天以上 interval 带 d 后缀，皆为前端原样
var LiqMapRanges = map[int]struct {
	Interval string
	Limit    int
}{
	1:   {"1", 1500},
	7:   {"5", 2000},
	30:  {"30", 1440},
	90:  {"90d", 1440},
	180: {"180d", 1440},
	365: {"365d", 1440},
}

// 🚪 LoginRequiredCode 服务端「需要登录」的业务码
// 📝 前端收到此码即弹出登录框；带了 obe 仍返回此码代表登录态失效
const LoginRequiredCode = "40000"

// 🔑 LoginHeaderName 登录凭证请求头名称
// 📝 前端拦截器读取 cookie「obe」并原样放入同名请求头
const LoginHeaderName = "obe"

// ⏱️ 请求配置
const (
	RequestTimeoutSeconds = 30 // HTTP请求超时时间（秒）

	// LoginMinRequestIntervalSeconds 登录状态下两次请求的最小间隔（秒）
	// 🛡️ 防封：带账号请求时自动限速，避免高频请求导致账号被风控
	LoginMinRequestIntervalSeconds = 5
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
