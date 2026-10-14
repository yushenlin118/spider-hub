# Coinglass 資料庫使用指南（給 AI agent）

PostgreSQL 裡 `coinglass_` 開頭的資料表（位於 `public` schema），由 `collectors/coinglass` 收集器定期從 coinglass.com 抓取寫入。內容包含 BTC 現貨 ETF 流入流出、恐懼貪婪指數、全球 M2、現貨／合約資金流入流出、清算地圖快照。

> **只讀使用**：分析時只做 `SELECT`。所有資料表都由收集器維護，不要寫入、刪除或修改結構。
>
> 同一個資料庫裡還有其他資料表，例如 freqtrade 交易機器人的 `trades`、`orders`，以及 `macro_*` 系列。它們**不是 coinglass 資料**，本指南不涵蓋，不要動它們。

## 1. 先讀資料庫註解

**每張表、每個欄位的意義、單位、歷史起點、使用陷阱，都寫在資料庫的註解（`COMMENT`）裡，那是唯一的說明來源。** 本指南只列欄位與型態（第 2 節），不重複說明，開始分析前請先查：

```sql
-- 所有 coinglass 表的表註解與欄位註解
SELECT c.relname AS table_name,
       a.attname AS column_name,
       COALESCE(col_description(c.oid, a.attnum), obj_description(c.oid)) AS description
FROM pg_class c
JOIN pg_namespace n ON n.oid = c.relnamespace
JOIN LATERAL (
  SELECT 0 AS attnum, '(table)'::name AS attname
  UNION ALL
  SELECT attnum, attname FROM pg_attribute
  WHERE attrelid = c.oid AND attnum > 0 AND NOT attisdropped
) a ON true
WHERE n.nspname = 'public' AND c.relname LIKE 'coinglass\_%' AND c.relkind = 'r'
ORDER BY c.relname, a.attnum;
```

使用 psql 時也可以用 `\d+ coinglass_coin_flow` 查單張表。

**表註解裡標示的歷史起點和陷阱一定要看**，尤其是：

- 有些資料**只有近期、而且永遠補不回更早的部分**（小時線、清算地圖），長期分析要改用註解建議的替代資料，不能用插值或其他粒度冒充。
- 有些表把多條序列放在一起，查詢時必須先過濾特定欄位，否則會把不同序列加在一起。

## 2. 資料表與欄位

欄位的意義、單位與陷阱見資料庫註解（上一節），這裡只列欄位與型態。

### `coinglass_etf_flow_daily`

主鍵：`date`

| 欄位 | 型態 |
|---|---|
| `date` | date |
| `net_flow_btc` | double precision |
| `net_flow_usd` | double precision |
| `updated_at` | timestamp with time zone |

### `coinglass_etf_flow_issuer_daily`

主鍵：`date, ticker`

| 欄位 | 型態 |
|---|---|
| `date` | date |
| `ticker` | text |
| `issuer` | text |
| `net_flow_btc` | double precision |
| `net_flow_usd` | double precision |
| `updated_at` | timestamp with time zone |

### `coinglass_etf_spot_inflow_daily`

主鍵：`date`

| 欄位 | 型態 |
|---|---|
| `date` | date |
| `net_flow_usd` | double precision |
| `net_assets_usd` | double precision |
| `btc_price_usd` | double precision |
| `updated_at` | timestamp with time zone |

### `coinglass_fear_greed_daily`

主鍵：`date`

| 欄位 | 型態 |
|---|---|
| `date` | date |
| `value` | double precision |
| `level` | text |
| `btc_price_usd` | double precision |
| `updated_at` | timestamp with time zone |

### `coinglass_global_m2_weekly`

主鍵：`week`

| 欄位 | 型態 |
|---|---|
| `week` | date |
| `m2_supply_usd` | double precision |
| `yoy_growth_pct` | double precision |
| `btc_price_usd` | double precision |
| `updated_at` | timestamp with time zone |

### `coinglass_coin_flow`

主鍵：`market, symbol, interval, bar_time`

| 欄位 | 型態 |
|---|---|
| `market` | text |
| `symbol` | text |
| `interval` | text |
| `bar_time` | timestamp with time zone |
| `inflow_usd` | double precision |
| `outflow_usd` | double precision |
| `net_flow_usd` | double precision |
| `is_final` | boolean |
| `updated_at` | timestamp with time zone |

### `coinglass_liq_map_snapshot`

主鍵：`id`

| 欄位 | 型態 |
|---|---|
| `id` | bigint |
| `captured_at` | timestamp with time zone |
| `map_type` | text |
| `query_symbol` | text |
| `range_days` | integer |
| `last_price` | double precision |
| `range_low` | double precision |
| `range_high` | double precision |

### `coinglass_liq_map_level`

主鍵：無

| 欄位 | 型態 |
|---|---|
| `snapshot_id` | bigint |
| `exchange` | text |
| `instrument_id` | text |
| `bucket` | double precision |
| `price` | double precision |
| `amount_usd` | double precision |
| `leverage` | integer |
| `tag` | text |

### `coinglass_collector_run`

主鍵：`id`

| 欄位 | 型態 |
|---|---|
| `id` | bigint |
| `job` | text |
| `started_at` | timestamp with time zone |
| `finished_at` | timestamp with time zone |
| `status` | text |
| `rows` | integer |
| `error` | text |
| `login_error` | boolean |

## 3. 通用規則

- 所有時間都是 **UTC**。`date` 欄位是 UTC 日期，`timestamptz` 欄位以 UTC 理解。
- 金額單位都是**美元**，除非欄位名稱帶 `_btc`。
- `updated_at` 是收集器最後一次寫入該列的時間，不是資料本身的時間。
- 時間序列表以主鍵 **upsert**：同一天／同一根 K 線被重抓時會覆蓋舊值，不會出現重複列。來源事後修正的數字也會被覆蓋，資料庫只保留最新值，不保留修改歷史。
- 清算地圖與執行紀錄是 **append**：每次新增，不覆蓋。

## 4. 確認資料新鮮度

分析前先確認收集器是否正常運作，避免拿過期資料下結論：

```sql
-- 每個任務最近一次執行結果
SELECT DISTINCT ON (job) job, started_at, status, rows, login_error, error
FROM coinglass_collector_run
ORDER BY job, started_at DESC;
```

- 各任務的頻率寫在 `coinglass_collector_run` 的表註解。最近一次 `started_at` 遠早於頻率，代表收集器可能停了。
- `login_error = true`：清算地圖的登入 token 失效，之後的清算地圖快照會缺漏，直到有人更新 token。

## 5. 確認實際的歷史範圍

註解裡的歷史起點是寫入當時的狀態。分析前以實際查詢為準，**使用者要求的期間如果超出範圍，要明確告知資料不存在**：

```sql
-- 各資料目前實際的歷史範圍
SELECT 'coin_flow ' || market || ' ' || interval AS data, min(bar_time)::date AS first, max(bar_time)::date AS last, count(*)
FROM coinglass_coin_flow GROUP BY market, interval
UNION ALL SELECT 'etf_flow_daily', min(date), max(date), count(*) FROM coinglass_etf_flow_daily
UNION ALL SELECT 'etf_spot_inflow_daily', min(date), max(date), count(*) FROM coinglass_etf_spot_inflow_daily
UNION ALL SELECT 'fear_greed_daily', min(date), max(date), count(*) FROM coinglass_fear_greed_daily
UNION ALL SELECT 'global_m2_weekly', min(week), max(week), count(*) FROM coinglass_global_m2_weekly
UNION ALL SELECT 'liq_map_snapshot ' || map_type, min(captured_at)::date, max(captured_at)::date, count(*)
FROM coinglass_liq_map_snapshot GROUP BY map_type
ORDER BY 1;

-- 清算地圖實際有快照的日期（檢查有沒有缺天）
SELECT captured_at::date AS day, count(*) AS snapshots
FROM coinglass_liq_map_snapshot GROUP BY 1 ORDER BY 1;
```

## 6. 查詢範例

```sql
-- 最近 30 天 BTC 現貨 vs 合約每日淨流入（只取已收盤）
SELECT bar_time::date AS day,
       sum(net_flow_usd) FILTER (WHERE market = 'spot')    / 1e6 AS spot_net_musd,
       sum(net_flow_usd) FILTER (WHERE market = 'futures') / 1e6 AS futures_net_musd
FROM coinglass_coin_flow
WHERE symbol = 'BTC' AND interval = 'h24' AND is_final
  AND bar_time >= now() - interval '30 days'
GROUP BY 1 ORDER BY 1;

-- ETF 淨流入與恐懼貪婪指數同日對照
SELECT e.date, e.net_flow_usd / 1e6 AS etf_net_musd, f.value AS fear_greed, f.level
FROM coinglass_etf_flow_daily e
JOIN coinglass_fear_greed_daily f USING (date)
ORDER BY e.date DESC LIMIT 30;

-- 最新一次清算地圖：現價上下各 5% 內的多空清算金額
WITH s AS (
  SELECT * FROM coinglass_liq_map_snapshot
  WHERE map_type = 'pair' AND query_symbol = 'Binance_BTCUSDT'
  ORDER BY captured_at DESC LIMIT 1
)
SELECT s.captured_at, s.last_price,
       sum(l.amount_usd) FILTER (WHERE l.bucket <  s.last_price) / 1e6 AS long_liq_musd,
       sum(l.amount_usd) FILTER (WHERE l.bucket >= s.last_price) / 1e6 AS short_liq_musd
FROM s JOIN coinglass_liq_map_level l ON l.snapshot_id = s.id
WHERE l.bucket BETWEEN s.last_price * 0.95 AND s.last_price * 1.05
GROUP BY s.captured_at, s.last_price;

-- 最新一次清算地圖：金額最大的 10 個價位（依交易所加總）
SELECT l.bucket, sum(l.amount_usd) / 1e6 AS amount_musd
FROM coinglass_liq_map_level l
WHERE l.snapshot_id = (SELECT id FROM coinglass_liq_map_snapshot
                       WHERE map_type = 'exchange' ORDER BY captured_at DESC LIMIT 1)
GROUP BY l.bucket ORDER BY amount_musd DESC LIMIT 10;

-- M2 年增率 vs BTC 價格（週）
SELECT week, yoy_growth_pct, btc_price_usd FROM coinglass_global_m2_weekly ORDER BY week;
```
