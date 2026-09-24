package coinglass

import "time"

// 📦 CoinglassResponse API原始响应结构
// 🔐 所有Coinglass API都返回加密的响应数据
// 📝 该结构定义在 spider.go 中，此处仅作说明，解密后的业务数据结构见下方各类型

// 🏦 ETFFlowDay 比特币现货ETF单日净流入流出
// 📊 对应接口: ETFFlowEndpoint (/api/etf/flow)
// 💡 Change 字段单位为BTC，对应前台图表标题的 (BTC)；ChangeUsd 为美元口径
type ETFFlowDay struct {
	Change    float64     `json:"change"`    // 📈 当日净流入流出（单位: BTC，负数为净流出）
	ChangeUsd float64     `json:"changeUsd"` // 💵 当日净流入流出（单位: 美元）
	Date      int64       `json:"date"`      // ⏰ 毫秒时间戳（UTC当日零点）
	List      []ETFIssuer `json:"list"`      // 🏢 各发行商当日明细
}

// 📅 Time 将毫秒时间戳转为UTC时间
// ⚠️ 必须以UTC解读，否则在东八区会整体偏移一天
func (d ETFFlowDay) Time() time.Time {
	return time.UnixMilli(d.Date).UTC()
}

// 🏢 ETFIssuer 单一ETF发行商的当日流入流出
type ETFIssuer struct {
	Change    float64 `json:"change"`    // 📈 该发行商当日净流入流出（单位: BTC）
	ChangeUsd float64 `json:"changeUsd"` // 💵 该发行商当日净流入流出（单位: 美元）
	Name      string  `json:"name"`      // 🏷️ 发行商名称，如 Blackrock
	Ticker    string  `json:"ticker"`    // 🔖 ETF代码，如 IBIT
}

// 📉 ETFSpotInFlowDay 比特币现货ETF单日净流入及资产快照（精简版，无发行商拆分）
// 📊 对应接口: ETFSpotInFlowEndpoint (/api/stock/spot/inFlow)
type ETFSpotInFlowDay struct {
	BtcPrice  float64 `json:"btcPrice"`  // 🪙 当日BTC价格（美元）
	Change    float64 `json:"change"`    // 💵 当日净流入流出（单位: 美元）
	Date      int64   `json:"date"`      // ⏰ 毫秒时间戳
	NetAssets float64 `json:"netAssets"` // 🏦 ETF总净资产（美元）
}

// 📅 Time 将毫秒时间戳转为UTC时间
func (d ETFSpotInFlowDay) Time() time.Time {
	return time.UnixMilli(d.Date).UTC()
}

// 📋 ETFInfo 比特币现货ETF基本信息
// 📊 对应接口: ETFListEndpoint (/api/bitcoin/etf/list)
type ETFInfo struct {
	Ticker      string  `json:"ticker"`      // 🔖 ETF代码，如 IBIT
	ETFName     string  `json:"etfName"`     // 🏷️ ETF全称
	Issuer      string  `json:"issuer"`      // 🏢 发行商
	Custodian   string  `json:"custodian"`   // 🔐 托管方，如 Coinbase
	Type        string  `json:"type"`        // 📑 类型，如 Spot
	Status      string  `json:"status"`      // ✅ 状态，如 Running
	Price       float64 `json:"price"`       // 💲 当前价格
	Aum         float64 `json:"aum"`         // 🏦 资产管理规模（美元）
	MarketCap   float64 `json:"marketCap"`   // 📊 市值（美元）
	Volume      float64 `json:"volume"`      // 🔄 成交额（美元）
	AvgDailyVol float64 `json:"avgDailyVol"` // 📈 日均成交量
	Fee         float64 `json:"fee"`         // 💸 管理费率，如 0.015 表示 1.5%
	FeeWaiver   float64 `json:"feeWaiver"`   // 🎁 费率减免
}

// 🔥 HeatmapCoin 币种热力图中的单个币种
// 📊 对应接口: CoinsHeatmapEndpoint (/api/futures/coins/heatmap)
// 💡 绘制热力图时: Value 决定方块大小，PriceChangePercent 决定方块颜色
type HeatmapCoin struct {
	Symbol             string  `json:"symbol"`             // 🪙 币种符号，如 BTC
	PriceChangePercent float64 `json:"priceChangePercent"` // 📈 24小时涨跌幅（百分比，如 1.02 表示 +1.02%）
	Value              float64 `json:"value"`              // 📊 指标值，含义由请求时的 type 参数决定
}
