# Coinglass 数据爬虫 🕷️

Coinglass 是一个专业的加密货币数据分析平台，提供期货、现货、持仓量等多维度的市场数据。

## 🏗️ 项目结构

## 功能特性 ✨

- 🔍 持仓量图表数据获取
- 📊 期货首页统计数据
- 🏢 衍生品交易所列表
- 💰 币种市场数据
- 🔗 交易所期货交易对信息
- 🪙 现货支持币种信息
- 🏦 比特币现货ETF净流入流出（含各发行商明细）
- 📋 比特币现货ETF清单（AUM、管理费率、托管方）
- 🔥 币种热力图（首页「热力图(24小时)」，1612个币种）

## API 接口说明 📋

### 1. 持仓量图表数据 (`/api/openInterest/v3/chart`) 📈
获取指定币种的持仓量图表数据

**请求示例:**
```
https://capi.coinglass.com/api/openInterest/v3/chart?symbol=BTC&timeType=0&exchangeName=&currency=USD&type=0
```
**实际测试结果:**
```
✅ 持仓量图表数据获取成功，数据长度: 422718
📊 响应时间: 3.81s
📄 数据预览: {"dataMap":{"Binance":[161745430,181309262,180784983,182010652,176903553,176218564,183792561,1963354...
```

### 2. 期货首页统计 (`/api/futures/home/statistics`) 🏠
获取期货市场的整体统计数据，无需参数

**请求示例:**
```
https://capi.coinglass.com/api/futures/home/statistics
```

**实际测试结果:**
```
✅ 期货首页统计数据获取成功，数据长度: 223
📊 响应时间: 0.35s
📄 数据预览: {"volH24Chain":-26.5900,"shortRate":51.8900,"oiH24Chain":1.5900,"openInterest":202194926751,"longRat...
```


### 3. 衍生品交易所列表 (`/api/derivative/exchange/list`) 🏢
获取支持衍生品交易的交易所列表，无需参数

**请求示例:**
```
https://capi.coinglass.com/api/derivative/exchange/list
```

**实际测试结果:**
```
✅ 衍生品交易所列表获取成功，数据长度: 7824
📊 响应时间: 0.24s
📄 数据预览: [{"exchangeName":"Binance","liquidationVolUsd":95900704.33530666,"logo":"https://cdn.coinglasscdn.co...
```


### 4. 币种市场数据 (`/api/home/v2/coinMarkets`) 💰
获取币种的市场数据和排行信息

**请求示例:**
```
https://capi.coinglass.com/api/home/v2/coinMarkets?sort=&order=&keyword=&pageNum=1&pageSize=20&ex=all
```

**参数说明:**
- `sort`: 排序字段 🔄
- `order`: 排序方向 ⬆️
- `keyword`: 搜索关键词 🔍
- `pageNum`: 页码 (默认: 1) 📄
- `pageSize`: 每页数量 (默认: 20) 📊
- `ex`: 交易所筛选 (all: 全部) 🏢

**实际测试结果:**
```
✅ 币种市场数据获取成功，数据长度: 53169
📊 响应时间: 0.81s
📄 数据预览: {"total":664,"pageSize":20,"list":[{"avgFundingRateByOi":0.011645,"avgFundingRateByOiAPR":12.7513,"a...
```

### 5. 交易所期货交易对信息 (`/api/exchange/futures/pairInfo`) 🔗
获取各交易所的期货交易对详细信息，无需参数

**请求示例:**
```
https://capi.coinglass.com/api/exchange/futures/pairInfo
```

**实际测试结果:**
```
✅ 交易所期货交易对信息获取成功，数据长度: 73495
📊 响应时间: 1.01s
📄 数据预览: [{"symbol":"BTC","symbolLogo":"https://cdn.coinglasscdn.com/static/img/coins/bitcoin-BTC.png"},{"sym...
```

### 6. 现货支持币种 (`/api/spot/support/coin`) 🪙
获取平台支持的现货交易币种列表，无需参数

**请求示例:**
```
https://capi.coinglass.com/api/spot/support/coin
```

**实际测试结果:**
```
✅ 现货支持币种获取成功，数据长度: 5794
📊 响应时间: 0.25s
📄 数据预览: ["BTC","ETH","XRP","USDT","BNB","SOL","USDC","DOGE","TRX","ADA","HYPE","SUI","XLM","LINK","HBAR","BC...
```

### 7. 币种市场数据（小数据量测试） (`/api/home/v2/coinMarkets`) 📊
获取少量币种数据用于快速测试

**请求示例:**
```
https://capi.coinglass.com/api/home/v2/coinMarkets?sort=&order=&keyword=&pageNum=1&pageSize=5&ex=all
```

**实际测试结果:**
```
✅ 市场数据获取成功，数据长度: 13592
📊 响应时间: 0.22s
📄 数据预览: {"total":664,"pageSize":5,"list":[{"avgFundingRateByOi":0.011645,"avgFundingRateByOiAPR":12.7513,"av...
```

### 8. 比特币现货ETF净流入流出 (`/api/etf/flow`) 🏦
获取比特币现货ETF每日净流入流出，对应前台图表「比特币现货ETF净流入流出 (BTC)」，无需参数

> 💡 **接口命名规律**：BTC接口不带币种前缀，其他币种才带，如 `/api/etf/eth/flow`、`/api/etf/sol/flow`、`/api/etf/xrp/flow`、`/api/etf/hk/flow`

**请求示例:**
```
https://capi.coinglass.com/api/etf/flow
```

**推荐用法（已封装类型化方法）:**
```go
spider := coinglass.NewSpider()

days, err := spider.GetETFFlow()
if err != nil {
    log.Fatal(err)
}

latest := days[len(days)-1] // 📅 最新一日
fmt.Printf("%s 净流入 %+.2f BTC (%+.1f 百万美元)\n",
    latest.Time().Format("2006-01-02"), latest.Change, latest.ChangeUsd/1e6)

for _, issuer := range latest.List { // 🏢 各发行商明细
    fmt.Printf("  %s %s: %+.2f BTC\n", issuer.Ticker, issuer.Name, issuer.Change)
}
```

**返回数据结构:**
```go
type ETFFlowDay struct {
    Change    float64     `json:"change"`    // 当日净流入流出（单位: BTC）
    ChangeUsd float64     `json:"changeUsd"` // 当日净流入流出（单位: 美元）
    Date      int64       `json:"date"`      // 毫秒时间戳（UTC当日零点）
    List      []ETFIssuer `json:"list"`      // 各发行商明细
}
```

**实际测试结果:**
```
✅ ETF净流入流出数据获取成功，共 693 天，区间 2024-01-11 ~ 2026-09-22

📅 最近5天比特币现货ETF净流入流出:
  2026-09-18      +5668.59 BTC  (   +433.0 百万美元)
  2026-09-21     +12311.45 BTC  (   +999.0 百万美元)
  2026-09-22      +8254.91 BTC  (   +714.7 百万美元)

🏢 2026-09-22 各发行商明细:
  IBIT   Blackrock            +4046.03 BTC     +350.3 百万美元
  FBTC   Fidelity             +2973.01 BTC     +257.4 百万美元
  MSBT   Morgan Stanley       +1143.47 BTC      +99.0 百万美元
```

**⚠️ 注意事项:**
- 该接口**不支持分页**，单次返回全量历史（约600KB），取最新一日用 `days[len(days)-1]`
- `Date` 为 **UTC** 当日零点，格式化时务必用 `.Time()`（内部已转UTC），否则在东八区会整体偏移一天
- 请勿与 `/api/etf/history/chart?type=2` 混淆，后者是**持仓存量/AUM**，不是每日净流入流出

### 9. 比特币现货ETF净流入（精简版） (`/api/stock/spot/inFlow`) 📉
只含每日净流入（美元）、ETF总净资产、BTC价格，无发行商拆分，适合只看总量趋势

**推荐用法:**
```go
days, err := spider.GetETFSpotInFlow()
latest := days[len(days)-1]
fmt.Printf("净流入 %+.1f 百万美元, 总净资产 %.1f 亿美元, BTC价格 $%.1f\n",
    latest.Change/1e6, latest.NetAssets/1e8, latest.BtcPrice)
```

**实际测试结果:**
```
✅ ETF净流入数据获取成功，共 676 天
📊 2026-09-22 快照: 净流入 +714.7 百万美元, 总净资产 1492.2 亿美元, BTC价格 $86158.6
```

### 10. 比特币现货ETF清单 (`/api/bitcoin/etf/list`) 📋
获取各ETF的代码、发行商、托管方、AUM、管理费率等基本信息，无需参数

**推荐用法:**
```go
list, err := spider.GetETFList()
for _, etf := range list {
    fmt.Printf("%s %s AUM: %.1f 亿美元 费率: %.2f%%\n",
        etf.Ticker, etf.Issuer, etf.Aum/1e8, etf.Fee*100)
}
```

**实际测试结果:**
```
✅ ETF清单获取成功，共 23 只

🏦 比特币现货ETF清单（前5只）:
  GBTC   Grayscale Bitcoin Trust        发行商: Grayscale    AUM: 286.2 亿美元  费率: 1.50%
  IBIT   iShares Bitcoin Trust          发行商: BlackRock    AUM: 7.1 亿美元  费率: 0.25%
  FBTC   Wise Origin Bitcoin Trust      发行商: Fidelity     AUM: 5.2 亿美元  费率: 0.25%
```

### 11. 币种热力图 (`/api/futures/coins/heatmap`) 🔥
获取首页「热力图(24小时)」数据，返回全部币种（实测 1612 个）的24小时涨跌幅与指标值

**请求示例:**
```
https://capi.coinglass.com/api/futures/coins/heatmap?type=vol
```

**参数说明:**
- `type`: 指标类型 🎨 **必填**，决定 `value` 字段的含义
  - `price` — `value` = 24小时涨跌幅
  - `vol` — `value` = 24小时成交额（美元）
  - `oi` — `value` = 持仓量（美元）

> ⚠️ `type` 只接受上述三个字符串。传数字（如 `type=1`）或其他值服务端返回 `{"code":"40001","msg":"Server Error"}`，且**不会**下发 `user` header，解密会失败。`GetCoinsHeatmap` 已在客户端提前拦截。

**推荐用法:**
```go
spider := coinglass.NewSpider()

coins, err := spider.GetCoinsHeatmap(coinglass.HeatmapTypeVol)
if err != nil {
    log.Fatal(err)
}

// 📐 Value 决定方块大小，PriceChangePercent 决定方块颜色
for _, c := range coins {
    fmt.Printf("%s: 成交额 %.2f 亿美元, 24h涨跌 %+.2f%%\n",
        c.Symbol, c.Value/1e8, c.PriceChangePercent)
}
```

**返回数据结构:**
```go
type HeatmapCoin struct {
    Symbol             string  `json:"symbol"`             // 币种符号
    PriceChangePercent float64 `json:"priceChangePercent"` // 24小时涨跌幅（%）
    Value              float64 `json:"value"`              // 指标值，含义由 type 决定
}
```

**实际测试结果:**
```
✅ 热力图获取成功 (type=vol, 24小时成交额)，共 1612 个币种

🔥 24小时成交额 排行（前5）:
  BTC            713.08 亿美元      24h涨跌   +0.96%
  ETH            499.64 亿美元      24h涨跌   +0.66%
  SOL             97.59 亿美元      24h涨跌   +1.44%
  XRP             95.02 亿美元      24h涨跌   +6.47%
  ZEC             81.48 亿美元      24h涨跌   +7.48%
```

**📌 其他热力图接口（均已实测，2026-09-23）:**

| 热力图 | 端点与参数 | 状态 |
|---|---|---|
| 清算热力图 | `/api/index/v2/liqHeatMap`（`merge` `symbol` `interval` `limit` `data`） | 🔒 **需要登录**，返回 `{"code":"40000"}` |
| 资金费率热力图 | `/api/fundingRate/heatmap?type=1&time=24h` | ⚠️ 解密正常，内容为 `null` |
| 交易所热力图 | `/api/derivative/exchange/heatmap?exName=Binance&type=1` | ⚠️ 返回727个币种，但 `value` 恒为 `0` |

> 后两个是在解密功能正常的前提下复测的，`null` 与 `value=0` 均为服务端真实响应，并非密钥问题导致的假象。

### 清算热力图的参数与签名 🔒

参数格式（从前端还原）：

| 参数 | 值 | 说明 |
|---|---|---|
| `symbol` | `Binance_BTCUSDT` | 格式为 `交易所_交易对` |
| `interval` | `h24` / `h12` / `d1` | ⚠️ 是 `h24` 不是 `24h` |
| `limit` | `288` | 默认值；`h12`→`360`、`d1`→`720` |
| `merge` | `true` | |
| `data` | TOTP签名 | 见下方 |

`data` 参数是一个 TOTP 签名，前端实现还原如下：

```js
// 1. 取 Unix 秒 t，用 TOTP 生成6位码（secret=I65VU7K5ZQL7WB4E, step=30）
// 2. 拼成 "<t>,<code>"
// 3. AES-256-ECB + Pkcs7 加密（key = "1f68efd73f8d4921acc0dead41dd39bc" 的UTF8字节）
// 4. base64 输出
```

签名本身已验证可被服务端接受（未带签名时报 `Required String parameter 'data' is not present`，带上后变为登录校验），**但该接口额外要求登录态**，因此本项目未实现该签名。若后续补上登录，按上述步骤生成 `data` 即可。

## 🧪 测试和示例

```bash
# 运行全部测试
cd platforms/coinglass && go test -v

# 只运行ETF相关测试
go test -run TestETF -v

# 只运行热力图测试
go test -run TestCoinsHeatmap -v

# 运行单个测试
go test -run TestETFFlow -v
```

## 🔑 密钥推导：响应头 `v` 决定来源

服务端会切换AES密钥的推导方式，通过响应头 `v` 告知客户端。`deriveKeySource()` 已实现全部分支：

| `v` | 密钥来源 |
|---|---|
| `0` | 请求时发送的 `cache-ts-v2`（毫秒时间戳） |
| `1` | **API路径**（从 `/api` 起，不含查询字符串），如 `/api/etf/flow` |
| `2` | 响应头 `time` |
| `55` / `66` / `77` | 固定值，见 `constants.go` 的 `KeySourceV55/66/77` |

取 `base64(来源)[:16]` 作为AES-128密钥解开 `user` header，得到gzip压缩的动态密钥，再用它解 `data`。

> 📌 **2026-09-23 变更**：服务端由 `v=0` 切换为 `v=1`。当时仅实现 `v=0`，导致全平台接口解密失败并报「无效的PKCS7填充」。现已支持全部分支，遇到未知 `v` 会明确报错。
>
> ⚠️ 响应头里还有个 `ev`（当前为 `2`），**它不是切换标志**——该头不在 `Access-Control-Expose-Headers` 名单中，前端读不到。认 `v`，不要认 `ev`。

**若服务端再次更换推导方式**，`deriveKeySource` 会返回「未知的加密版本」错误。还原方法：前端 `app-*.js` 中的 interceptor 采用字符串表混淆，把解码器 `Mn`（别名 `bn`）、字符串数组 `wn` 与 checksum 为 `191875` 的旋转IIFE一并抽出到 node 中执行，即可还原 `bn(341,"AQ3U")` 这类调用。密钥推导位于 `mn()`。

验证技巧：密钥正确时 `user` header 解密后是gzip（magic `1f 8b`）。以此为判据可**离线**穷举密钥，无需反复请求线上接口。

## ⚠️ 免责声明

### 法律声明 📋
本项目仅供学习和研究使用，请遵守相关网站的使用条款和robots.txt协议。使用者需要自行承担使用风险，作者不承担任何法律责任。

1. **🔍 学习目的**：本爬虫工具仅用于技术学习、数据分析研究等合法用途
2. **📚 教育用途**：适用于编程教学、算法研究、数据科学学习等教育场景
3. **🚫 禁止滥用**：严禁用于恶意攻击、数据盗取、商业牟利等非法行为
4. **⚖️ 法律责任**：使用者需自行承担因违法使用而产生的一切法律后果

### 使用限制 🚦

1. **📊 数据使用**：
   - 获取的数据仅供个人学习和研究使用
   - 不得将数据用于商业目的或二次销售
   - 不得恶意大量请求导致服务器负载过重

2. **🔐 隐私保护**：
   - 尊重目标网站的robots.txt协议
   - 不得获取用户隐私信息或敏感数据
   - 遵守相关数据保护法规

3. **⏱️ 访问频率**：
   - 合理控制请求频率，避免对目标服务器造成压力
   - 建议在请求间设置适当的延时
   - 遵守网站的访问限制和使用条款

### 风险提示 ⚡

1. **🔄 数据变化**：目标网站可能随时更改API接口或数据结构
2. **🛡️ 反爬措施**：网站可能采用反爬虫技术，导致爬虫失效
3. **📈 市场风险**：获取的金融数据仅供参考，不构成投资建议
4. **⚖️ 法律风险**：请确保使用行为符合当地法律法规

### 版权说明 📄

- 本项目代码采用 [Apache 2.0](../../LICENSE) 开源协议，详见 LICENSE 文件
- Coinglass 是 Coinglass 公司的注册商标，本项目与其无任何官方关联
- 所有数据版权归 Coinglass 平台所有
- 项目中使用的Logo、商标等知识产权归原权利人所有
- 更多详细信息请访问 [Coinglass 官网](https://www.coinglass.com/) 🌐
---

**⚡ Spider-Hub**

> 如果您觉得这个项目有用，请给我们一个 ⭐ Star，这是对我们最大的鼓励！

