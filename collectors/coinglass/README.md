# Coinglass 收集器 🗄️

定期呼叫 `platforms/coinglass` 抓取資料，寫入 PostgreSQL。資料表與欄位的說明寫在資料庫註解裡（來源是 `schema.sql` 尾段的 `COMMENT ON`）；[DATA.md](DATA.md) 是給 AI agent 的通用使用指南，教它怎麼查註解、確認資料新鮮度與歷史範圍。

## 收集內容與頻率

| 任務 | 資料 | 頻率 | 登入 |
|---|---|---|---|
| `etf_flow` | BTC 現貨 ETF 每日流入流出（含發行商） | 6 小時 | 否 |
| `etf_spot_inflow` | BTC 現貨 ETF 淨流入 + AUM | 6 小時 | 否 |
| `fear_greed` | 恐懼貪婪指數 | 6 小時 | 否 |
| `global_m2` | 全球 M2 | 24 小時 | 否 |
| `coin_flow` | BTC 現貨／合約流入流出（h1、h24） | 1 小時 | 否 |
| `liq_map` | BTC 清算地圖快照（交易對 + 交易所） | 24 小時 | 🔒 是 |

首次執行會自動回補伺服端給得到的全部歷史，之後只抓最近一段增量。

## 環境變數

| 變數 | 必填 | 說明 |
|---|---|---|
| `DATABASE_URL` | ✅ | `postgres://user:pass@host:5432/db?sslmode=...` |
| `COINGLASS_OBE` | | 瀏覽器 cookie `obe` 的值；不設定就跳過 `liq_map` |
| `COINGLASS_UA` | | 取得 `obe` 的瀏覽器 User-Agent，與登入狀態保持一致 |

啟動時會自動在既有資料庫的 `public` schema 建立資料表，並寫入每張表與每個欄位的註解（`COMMENT`）（`schema.sql`，全用 `IF NOT EXISTS`，可重複執行）。表名與索引名一律加 `coinglass_` 前綴，不會建立 schema 或資料庫，也不會碰其他資料表。

## 執行

```bash
# 常駐模式：每分鐘檢查排程，到期的任務依序執行
go run ./collectors/coinglass

# 全部任務立刻跑一次就結束（手動補抓、驗證用）
go run ./collectors/coinglass -once

# 只跑指定任務
go run ./collectors/coinglass -once -job coin_flow,fear_greed

# 只套用 schema.sql（建表、更新資料表與欄位註解），不抓資料
go run ./collectors/coinglass -migrate
```

## Docker 部署

在 **repo 根目錄**建置（需要引用 `platforms/coinglass`）：

```bash
docker build -f collectors/coinglass/Dockerfile -t coinglass-collector .
docker run -d --name coinglass-collector --restart unless-stopped \
  -e DATABASE_URL='postgres://...' \
  -e COINGLASS_OBE='...' \
  coinglass-collector
```

## 排程機制

- 是否到期是看資料庫 `coinglass_collector_run` 裡該任務最近一次的紀錄，所以**重啟不會重複抓**，停機一段時間後重啟也會立刻補跑到期的任務。
- 成功：距上次滿 `Every` 才再跑。
- 一般失敗（網路、伺服端錯誤）：過 `RetryAfter`（10 分鐘～1 小時）重試。
- 登入失效：等滿 `Every`（24 小時）才重試，不會一直用失效的 token 打伺服端，降低帳號被封的風險。
- token 只用在清算地圖：公開任務用另一個不帶帳號的爬蟲，不會用你的帳號發請求，也不受登入限速影響。清算地圖的請求之間至少間隔 5 秒。

## 維運

```sql
-- 各任務最近狀態
SELECT DISTINCT ON (job) job, started_at, status, rows, login_error, error
FROM coinglass_collector_run ORDER BY job, started_at DESC;
```

- **`liq_map` 出現 `login_error = true`**：token 過期了。從瀏覽器重新取得 `obe`，更新 `COINGLASS_OBE` 後重啟容器。重啟後會立刻補跑一次。
- 頻率、幣種、粒度都定義在 `jobs.go` 開頭（`Jobs`、`flowSymbols`、`flowIntervals`、`liqTargets`）。
- 流入流出的增量筆數會依資料庫裡最新一根 K 線距今多久自動計算（至少 48 根），停機後重啟會自動補齊。唯一的限制是伺服端只保留小時線約 180 天（4320 根），**停機超過 180 天**，中間的小時線就補不回來了。
