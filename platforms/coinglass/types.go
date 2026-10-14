package coinglass

import (
	"encoding/json"
	"fmt"
	"time"
)

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

// 😱 FearGreedPoint 恐惧贪婪指数单日数据
// 📊 对应接口: FearGreedIndexEndpoint (/api/index/history) 的 data[0]
// 📝 原始响应为 dates/values/prices 三个等长数组，由 parseFearGreedIndex 按下标合并
type FearGreedPoint struct {
	Date  int64   // ⏰ 毫秒时间戳
	Value float64 // 🌡️ 指数值（0~100）
	Price float64 // 🪙 当日BTC价格（美元），服务端返回 null 时为 0
}

// 📅 Time 将毫秒时间戳转为UTC时间
func (p FearGreedPoint) Time() time.Time {
	return time.UnixMilli(p.Date).UTC()
}

// 🏷️ Level 返回指数所属等级（分界同前端，见 FearGreed*Max 常量）
func (p FearGreedPoint) Level() string {
	switch {
	case p.Value <= FearGreedExtremeFearMax:
		return "极度恐惧"
	case p.Value <= FearGreedFearMax:
		return "恐惧"
	case p.Value <= FearGreedNeutralMax:
		return "中性"
	case p.Value <= FearGreedGreedMax:
		return "贪婪"
	default:
		return "极度贪婪"
	}
}

// 💵 M2SupplyGrowthPoint 比特币 vs M2货币供应单周数据
// 📊 对应接口: GlobalM2SupplyGrowthEndpoint (/api/index/globalM2SupplyGrowth)
// 💡 前端画法: YoYGrowth 为柱状（正绿负红），M2Supply 与 Price 为折线（Price 用对数轴）
type M2SupplyGrowthPoint struct {
	Timestamp int64   `json:"timestamp"` // ⏰ 毫秒时间戳（每周一笔）
	Price     float64 `json:"price"`     // 🪙 BTC价格（美元）
	M2Supply  float64 `json:"m2Supply"`  // 💵 M2货币供应量（美元）
	YoYGrowth float64 `json:"yoyGrowth"` // 📈 M2同比增长率（百分比，如 8.0 表示 +8.0%）
}

// 📅 Time 将毫秒时间戳转为UTC时间
func (p M2SupplyGrowthPoint) Time() time.Time {
	return time.UnixMilli(p.Timestamp).UTC()
}

// 💸 FlowBar 单根资金流入流出K线
// 📊 对应接口: FlowKlineEndpoint (/api/v2/kline)
// 📝 原始为三元数组 [秒级时间戳, 流入USD, 流出USD]
type FlowBar struct {
	Timestamp  int64   // ⏰ 秒级时间戳（K线开盘时间）
	InflowUSD  float64 // 🟢 流入（主动买入成交额，美元）
	OutflowUSD float64 // 🔴 流出（主动卖出成交额，美元）
}

// 📅 Time 将秒级时间戳转为UTC时间
// ⚠️ 注意这里是秒，不是其他接口常见的毫秒
func (b FlowBar) Time() time.Time {
	return time.Unix(b.Timestamp, 0).UTC()
}

// 📈 NetFlowUSD 净流入（流入 - 流出），与前端「净流入」开关算法相同
func (b FlowBar) NetFlowUSD() float64 {
	return b.InflowUSD - b.OutflowUSD
}

// UnmarshalJSON 解析 [时间戳, 流入, 流出] 三元数组
func (b *FlowBar) UnmarshalJSON(data []byte) error {
	var arr []float64
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}
	if len(arr) != 3 {
		return fmt.Errorf("❌ 流入流出K线应为3个元素，实际 %d 个: %s", len(arr), data)
	}
	b.Timestamp, b.InflowUSD, b.OutflowUSD = int64(arr[0]), arr[1], arr[2]
	return nil
}

// 💥 LiquidationMap 单一交易对清算地图
// 📊 对应接口: LiqMapEndpoint (/api/index/5/liqMap)
// 💡 前端以 Leverage 分组、以 Bucket 为横轴画柱状图，并从 LastPrice 起累加画累计曲线
type LiquidationMap struct {
	Instrument LiqInstrument // 🔗 交易对信息
	LastPrice  float64       // 💲 当前价格
	Levels     []LiqLevel    // 📊 清算档位，按 Bucket 升序
}

// 🔗 LiqInstrument 清算地图的交易对信息
type LiqInstrument struct {
	ExName          string  `json:"exName"`          // 🏢 交易所，如 Binance
	InstrumentID    string  `json:"instrumentId"`    // 🆔 交易对，如 BTCUSDT
	BaseAsset       string  `json:"baseAsset"`       // 🪙 基础币种，如 BTC
	QuoteAsset      string  `json:"quoteAsset"`      // 💵 计价币种，如 USDT
	ContractType    string  `json:"contractType"`    // 📜 合约类型，如 PERPETUAL
	MaxLeverage     int     `json:"maxLeverage"`     // ⚡ 最大杠杆
	FundingInterval int     `json:"fundingInterval"` // ⏱️ 资金费率结算间隔（小时）
	PricePrecision  int     `json:"pricePrecision"`  // 🎯 价格精度（小数位）
	PriceTick       float64 `json:"priceTick"`       // 📏 最小价格变动
}

// 🏢 ExchangeLiquidationMap 交易所聚合清算地图
// 📊 对应接口: ExLiqMapEndpoint (/api/index/2/exLiqMap)
// 💡 前端以交易所分色画柱状图；现价以下累加为多单清算曲线，现价以上累加为空单清算曲线
type ExchangeLiquidationMap struct {
	Instrument LiqInstrument    // 🔗 参考交易对（实测为 Binance BTCUSDT，现价即取自它）
	LastPrice  float64          // 💲 当前价格
	RangeHigh  float64          // ⬆️ 价格范围上限
	RangeLow   float64          // ⬇️ 价格范围下限
	Exchanges  []ExchangeLiqMap // 🏢 各交易所数据，实测为 Binance / OKX / Bybit
}

// 🏢 ExchangeLiqMap 单一交易所的清算档位
// ⚠️ 聚合版档位的 Leverage 为 0、Tag 为空（服务端返回 null），每个区间只有一笔
type ExchangeLiqMap struct {
	Instrument LiqInstrument // 🔗 该交易所的交易对，如 OKX BTC-USDT-SWAP
	Levels     []LiqLevel    // 📊 清算档位，按 Bucket 升序
}

// 📊 LiqLevel 单个清算档位
// 📝 原始格式为四元数组 [价格, 金额, 杠杆, 标签]，挂在价格区间 key 之下
type LiqLevel struct {
	Bucket    float64 // 🪣 价格区间（原始 map 的 key，恒 >= Price，前端以此为横轴）
	Price     float64 // 💲 精确清算价格（前端读取但未使用）
	AmountUSD float64 // 💰 清算金额（美元）
	Leverage  int     // ⚡ 杠杆倍数，实测为 10 / 25 / 50 / 100（交易所聚合版为 0）
	Tag       string  // 🏷️ h1 / h2 / h3，含义未知：前端完全未使用，三者金额分布也无明显差异（聚合版为空）
}

// UnmarshalJSON 解析 [价格, 金额, 杠杆, 标签] 四元数组
// 📝 Bucket 不在数组内，由 GetLiquidationMap 从 map key 回填
func (l *LiqLevel) UnmarshalJSON(b []byte) error {
	var arr []json.RawMessage
	if err := json.Unmarshal(b, &arr); err != nil {
		return err
	}
	if len(arr) != 4 {
		return fmt.Errorf("❌ 清算档位应为4个元素，实际 %d 个: %s", len(arr), b)
	}

	if err := json.Unmarshal(arr[0], &l.Price); err != nil {
		return fmt.Errorf("❌ 解析清算价格失败: %w", err)
	}
	if err := json.Unmarshal(arr[1], &l.AmountUSD); err != nil {
		return fmt.Errorf("❌ 解析清算金额失败: %w", err)
	}
	if err := json.Unmarshal(arr[2], &l.Leverage); err != nil {
		return fmt.Errorf("❌ 解析杠杆倍数失败: %w", err)
	}
	if err := json.Unmarshal(arr[3], &l.Tag); err != nil {
		return fmt.Errorf("❌ 解析标签失败: %w", err)
	}
	return nil
}
