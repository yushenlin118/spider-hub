-- 🗄️ Coinglass 收集器资料表
-- 📝 程序启动时自动执行，全部使用 IF NOT EXISTS，可重复执行
-- ⚠️ 与 freqtrade 共用同一个资料库的 public schema，表名与索引名一律加 coinglass_ 前缀，
--    只会新增这些表，不会建立 schema，也不会碰其他资料表
-- 📖 资料表与栏位的说明以本档尾段的 COMMENT ON 为唯一来源（DATA.md 只是通用使用指南，不重复栏位说明）

-- 🏦 比特币现货ETF每日净流入流出（总量）
CREATE TABLE IF NOT EXISTS public.coinglass_etf_flow_daily (
    date          date PRIMARY KEY,                 -- 📅 交易日（UTC）
    net_flow_btc  double precision NOT NULL,        -- 📈 当日净流入流出（BTC，负数为净流出）
    net_flow_usd  double precision NOT NULL,        -- 💵 当日净流入流出（美元）
    updated_at    timestamptz NOT NULL DEFAULT now()
);

-- 🏢 比特币现货ETF每日净流入流出（各发行商明细）
CREATE TABLE IF NOT EXISTS public.coinglass_etf_flow_issuer_daily (
    date          date NOT NULL,
    ticker        text NOT NULL,                    -- 🔖 ETF代码，如 IBIT
    issuer        text NOT NULL,                    -- 🏷️ 发行商，如 Blackrock
    net_flow_btc  double precision NOT NULL,
    net_flow_usd  double precision NOT NULL,
    updated_at    timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (date, ticker)
);

-- 📉 比特币现货ETF每日净流入 + 总净资产 + 币价
CREATE TABLE IF NOT EXISTS public.coinglass_etf_spot_inflow_daily (
    date            date PRIMARY KEY,
    net_flow_usd    double precision NOT NULL,      -- 💵 当日净流入流出（美元）
    net_assets_usd  double precision NOT NULL,      -- 🏦 ETF总净资产（美元）
    btc_price_usd   double precision NOT NULL,      -- 🪙 当日BTC价格
    updated_at      timestamptz NOT NULL DEFAULT now()
);

-- 😱 恐惧贪婪指数（每日）
CREATE TABLE IF NOT EXISTS public.coinglass_fear_greed_daily (
    date           date PRIMARY KEY,
    value          double precision NOT NULL,       -- 🌡️ 0~100
    level          text NOT NULL,                   -- 🏷️ 极度恐惧/恐惧/中性/贪婪/极度贪婪
    btc_price_usd  double precision,                -- 🪙 服务端偶尔返回 null
    updated_at     timestamptz NOT NULL DEFAULT now()
);

-- 💵 全球M2货币供应（每周）
CREATE TABLE IF NOT EXISTS public.coinglass_global_m2_weekly (
    week            date PRIMARY KEY,               -- 📅 该周起始日（UTC）
    m2_supply_usd   double precision NOT NULL,      -- 💵 全球M2（美元）
    yoy_growth_pct  double precision NOT NULL,      -- 📈 同比增长（%）
    btc_price_usd   double precision NOT NULL,
    updated_at      timestamptz NOT NULL DEFAULT now()
);

-- 💸 现货 / 合约资金流入流出K线
CREATE TABLE IF NOT EXISTS public.coinglass_coin_flow (
    market        text NOT NULL CHECK (market IN ('spot', 'futures')),
    symbol        text NOT NULL,                    -- 🪙 如 BTC
    interval      text NOT NULL,                    -- ⏱️ h1 / h24
    bar_time      timestamptz NOT NULL,             -- ⏰ K线开盘时间
    inflow_usd    double precision NOT NULL,        -- 🟢 主动买入成交额
    outflow_usd   double precision NOT NULL,        -- 🔴 主动卖出成交额
    net_flow_usd  double precision GENERATED ALWAYS AS (inflow_usd - outflow_usd) STORED,
    is_final      boolean NOT NULL,                 -- ✅ 写入时该K线是否已收盘；false 表示仍在进行中，之后会被覆盖
    updated_at    timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (market, symbol, interval, bar_time)
);

-- 💥 清算地图快照（每次抓取一笔）
CREATE TABLE IF NOT EXISTS public.coinglass_liq_map_snapshot (
    id            bigserial PRIMARY KEY,
    captured_at   timestamptz NOT NULL DEFAULT now(),
    map_type      text NOT NULL CHECK (map_type IN ('pair', 'exchange')),
    query_symbol  text NOT NULL,                    -- 🪙 pair 为 Binance_BTCUSDT，exchange 为 BTC
    range_days    int NOT NULL,                     -- 📅 页面的时间范围选项
    last_price    double precision NOT NULL,        -- 💲 抓取当下现价
    range_low     double precision,                 -- ⬇️ 仅 exchange 有
    range_high    double precision                  -- ⬆️ 仅 exchange 有
);
CREATE INDEX IF NOT EXISTS coinglass_liq_map_snapshot_lookup
    ON public.coinglass_liq_map_snapshot (map_type, query_symbol, captured_at DESC);

-- 📊 清算地图档位
CREATE TABLE IF NOT EXISTS public.coinglass_liq_map_level (
    snapshot_id    bigint NOT NULL REFERENCES public.coinglass_liq_map_snapshot (id) ON DELETE CASCADE,
    exchange       text NOT NULL,                   -- 🏢 Binance / OKX / Bybit
    instrument_id  text NOT NULL,                   -- 🔗 如 BTCUSDT、BTC-USDT-SWAP
    bucket         double precision NOT NULL,       -- 🪣 价格区间（前端横轴）
    price          double precision NOT NULL,       -- 💲 精确清算价格
    amount_usd     double precision NOT NULL,       -- 💰 清算金额
    leverage       int,                             -- ⚡ 10/25/50/100；exchange 类型为 NULL
    tag            text                             -- 🏷️ h1/h2/h3，含义未知；exchange 类型为 NULL
);
CREATE INDEX IF NOT EXISTS coinglass_liq_map_level_snapshot ON public.coinglass_liq_map_level (snapshot_id);

-- 📋 收集器执行纪录（同时作为排程的依据，重启后不会重复抓）
CREATE TABLE IF NOT EXISTS public.coinglass_collector_run (
    id           bigserial PRIMARY KEY,
    job          text NOT NULL,
    started_at   timestamptz NOT NULL,
    finished_at  timestamptz NOT NULL,
    status       text NOT NULL CHECK (status IN ('ok', 'error')),
    rows         int NOT NULL DEFAULT 0,          -- 📏 本次写入（含更新）的笔数
    error        text,
    login_error  boolean NOT NULL DEFAULT false   -- 🚪 是否为登录失效（obe 过期）
);
CREATE INDEX IF NOT EXISTS coinglass_collector_run_job ON public.coinglass_collector_run (job, started_at DESC);

-- 📝 资料表与栏位注解（写进资料库，供直接连线的 AI agent / 分析者用 \d+ 或 obj_description 查看）
-- 🔁 COMMENT ON 可重复执行，每次启动都会套用最新内容；改表或改说明只需改这里

-- 🏦 coinglass_etf_flow_daily
COMMENT ON TABLE public.coinglass_etf_flow_daily IS
    'Coinglass：美國比特幣現貨 ETF 每日淨流入流出（所有 ETF 合計）。每個美股交易日一筆，週末與美國假日沒有列（不是缺漏）。自 2024-01-11（ETF 上市首日）起，已是完整歷史。最新一筆通常是前一個交易日。收集器每 6 小時 upsert。';
COMMENT ON COLUMN public.coinglass_etf_flow_daily.date IS '交易日（UTC 日期，主鍵）';
COMMENT ON COLUMN public.coinglass_etf_flow_daily.net_flow_btc IS '當日所有 BTC 現貨 ETF 合計淨流入，單位 BTC；負數為淨流出';
COMMENT ON COLUMN public.coinglass_etf_flow_daily.net_flow_usd IS '當日所有 BTC 現貨 ETF 合計淨流入，單位美元；負數為淨流出';
COMMENT ON COLUMN public.coinglass_etf_flow_daily.updated_at IS '收集器最後一次寫入此列的時間（不是資料本身的時間）';

-- 🏢 coinglass_etf_flow_issuer_daily
COMMENT ON TABLE public.coinglass_etf_flow_issuer_daily IS
    'Coinglass：美國比特幣現貨 ETF 每日淨流入流出的各 ETF 明細。主鍵 (date, ticker)。同日各 ticker 加總通常等於 coinglass_etf_flow_daily，但少數日子來源資料本身不一致（2026-10 檢查時 701 天中有 10 天），要用總量時以 coinglass_etf_flow_daily 為準。';
COMMENT ON COLUMN public.coinglass_etf_flow_issuer_daily.date IS '交易日（UTC 日期）';
COMMENT ON COLUMN public.coinglass_etf_flow_issuer_daily.ticker IS 'ETF 代碼，如 IBIT、FBTC、GBTC';
COMMENT ON COLUMN public.coinglass_etf_flow_issuer_daily.issuer IS '發行商名稱，如 Blackrock';
COMMENT ON COLUMN public.coinglass_etf_flow_issuer_daily.net_flow_btc IS '該 ETF 當日淨流入，單位 BTC；負數為淨流出';
COMMENT ON COLUMN public.coinglass_etf_flow_issuer_daily.net_flow_usd IS '該 ETF 當日淨流入，單位美元；負數為淨流出';
COMMENT ON COLUMN public.coinglass_etf_flow_issuer_daily.updated_at IS '收集器最後一次寫入此列的時間（不是資料本身的時間）';

-- 📉 coinglass_etf_spot_inflow_daily
COMMENT ON TABLE public.coinglass_etf_spot_inflow_daily IS
    'Coinglass：美國比特幣現貨 ETF 每日淨流入、總淨資產（AUM）與 BTC 價格。與 coinglass_etf_flow_daily 是不同來源接口，筆數與數值可能略有差異；要看 AUM 用這張表，要看發行商明細用 coinglass_etf_flow_issuer_daily。自 2024-01-11 起。';
COMMENT ON COLUMN public.coinglass_etf_spot_inflow_daily.date IS '交易日（UTC 日期，主鍵）';
COMMENT ON COLUMN public.coinglass_etf_spot_inflow_daily.net_flow_usd IS '當日淨流入，單位美元；負數為淨流出';
COMMENT ON COLUMN public.coinglass_etf_spot_inflow_daily.net_assets_usd IS '所有 BTC 現貨 ETF 的總淨資產（AUM），單位美元';
COMMENT ON COLUMN public.coinglass_etf_spot_inflow_daily.btc_price_usd IS '當日 BTC 價格（美元）';
COMMENT ON COLUMN public.coinglass_etf_spot_inflow_daily.updated_at IS '收集器最後一次寫入此列的時間（不是資料本身的時間）';

-- 😱 coinglass_fear_greed_daily
COMMENT ON TABLE public.coinglass_fear_greed_daily IS
    'Coinglass：加密貨幣恐懼貪婪指數，每天一筆，自 2018-02-01 起，已是完整歷史。收集器每 6 小時 upsert。';
COMMENT ON COLUMN public.coinglass_fear_greed_daily.date IS '日期（UTC，主鍵）';
COMMENT ON COLUMN public.coinglass_fear_greed_daily.value IS '指數 0～100，越低越恐懼';
COMMENT ON COLUMN public.coinglass_fear_greed_daily.level IS '等級，簡體中文字串：极度恐惧（≤20）、恐惧（≤40）、中性（≤60）、贪婪（≤80）、极度贪婪（>80）；分界與 coinglass 頁面一致，篩選時字串須完全相同';
COMMENT ON COLUMN public.coinglass_fear_greed_daily.btc_price_usd IS '當日 BTC 價格（美元）；來源偶爾不提供，此時為 NULL';
COMMENT ON COLUMN public.coinglass_fear_greed_daily.updated_at IS '收集器最後一次寫入此列的時間（不是資料本身的時間）';

-- 💵 coinglass_global_m2_weekly
COMMENT ON TABLE public.coinglass_global_m2_weekly IS
    'Coinglass：全球 M2 貨幣供應與 BTC 價格，每週一筆，自 2013-05-20 起，已是完整歷史。M2 是總體經濟統計，最新一筆通常落後今天約一個月，不要拿最新一筆 M2 直接對比今天的價格。來源會回頭修正舊數字，此表只保留最新修正值。';
COMMENT ON COLUMN public.coinglass_global_m2_weekly.week IS '該週日期（UTC，主鍵）';
COMMENT ON COLUMN public.coinglass_global_m2_weekly.m2_supply_usd IS '全球 M2 總量，單位美元';
COMMENT ON COLUMN public.coinglass_global_m2_weekly.yoy_growth_pct IS 'M2 年增率，單位百分點（8.0 代表 +8.0%）';
COMMENT ON COLUMN public.coinglass_global_m2_weekly.btc_price_usd IS '當週 BTC 價格（美元）';
COMMENT ON COLUMN public.coinglass_global_m2_weekly.updated_at IS '收集器最後一次寫入此列的時間（不是資料本身的時間）';

-- 💸 coinglass_coin_flow
COMMENT ON TABLE public.coinglass_coin_flow IS
    'Coinglass：現貨／合約資金流入流出 K 線，彙總各交易所主動買入（流入）與主動賣出（流出）成交額；衡量買賣力道，不是交易所間轉帳或鏈上流動。四條序列混在同一張表，查詢務必先過濾 market 與 interval。歷史起點：現貨 h24 2017-08-17、合約 h24 2019-09-08；h1 只從 2026-04-08 起，因為 coinglass 只保留約 180 天小時線，永遠沒有更早的 h1，長期分析請用 h24，不要用日線冒充或插值出小時資料。合約量約為現貨 10 倍以上，比較兩市場時請標準化。收集器每小時 upsert。';
COMMENT ON COLUMN public.coinglass_coin_flow.market IS '市場：spot（現貨）或 futures（合約）';
COMMENT ON COLUMN public.coinglass_coin_flow.symbol IS '幣種，目前只有 BTC';
COMMENT ON COLUMN public.coinglass_coin_flow.interval IS 'K 線粒度：h1（小時線）或 h24（日線）。兩者由來源分別提供，24 根 h1 加總不一定等於當天 h24';
COMMENT ON COLUMN public.coinglass_coin_flow.bar_time IS 'K 線開盤時間；h24 為 UTC 00:00';
COMMENT ON COLUMN public.coinglass_coin_flow.inflow_usd IS '流入：主動買入成交額（美元）';
COMMENT ON COLUMN public.coinglass_coin_flow.outflow_usd IS '流出：主動賣出成交額（美元）';
COMMENT ON COLUMN public.coinglass_coin_flow.net_flow_usd IS '淨流入 = inflow_usd - outflow_usd（自動計算欄位，美元）';
COMMENT ON COLUMN public.coinglass_coin_flow.is_final IS '寫入時此 K 線是否已收盤。false 表示仍在進行中，數值會隨時間變動並被覆蓋；計算平均、加總、比較時應加 WHERE is_final';
COMMENT ON COLUMN public.coinglass_coin_flow.updated_at IS '收集器最後一次寫入此列的時間（不是資料本身的時間）';

-- 💥 coinglass_liq_map_snapshot
COMMENT ON TABLE public.coinglass_liq_map_snapshot IS
    'Coinglass：BTC 清算地圖快照，一次抓取一列，append 寫入。內容是各價位「預估」會被清算的槓桿部位金額，不是已發生的清算。網站沒有歷史快照，資料只從 2026-10-05 起每天累積，收集器沒執行的日子會永久缺漏（例如 2026-10-07），做比較前先查實際有哪些日期。各價位明細在 coinglass_liq_map_level。';
COMMENT ON COLUMN public.coinglass_liq_map_snapshot.id IS '快照 ID，對應 coinglass_liq_map_level.snapshot_id';
COMMENT ON COLUMN public.coinglass_liq_map_snapshot.captured_at IS '抓取時間；代表當下那一刻的狀態，不代表整天';
COMMENT ON COLUMN public.coinglass_liq_map_snapshot.map_type IS 'pair：單一交易對，各價位依槓桿拆分；exchange：依交易所拆分（Binance、OKX、Bybit），無槓桿資訊';
COMMENT ON COLUMN public.coinglass_liq_map_snapshot.query_symbol IS '查詢對象：pair 為 Binance_BTCUSDT，exchange 為 BTC';
COMMENT ON COLUMN public.coinglass_liq_map_snapshot.range_days IS 'coinglass 頁面的時間範圍選項（天），目前固定為 1';
COMMENT ON COLUMN public.coinglass_liq_map_snapshot.last_price IS '抓取當下的 BTC 價格（美元）；bucket 低於此價為多單清算區，高於此價為空單清算區';
COMMENT ON COLUMN public.coinglass_liq_map_snapshot.range_low IS '價格範圍下限（美元）；只有 exchange 類型有值';
COMMENT ON COLUMN public.coinglass_liq_map_snapshot.range_high IS '價格範圍上限（美元）；只有 exchange 類型有值';

-- 📊 coinglass_liq_map_level
COMMENT ON TABLE public.coinglass_liq_map_level IS
    'Coinglass：清算地圖快照的各價位明細，一個快照約 120～380 列。pair 類型同一價位會依槓桿拆成多列，分析價位分佈時依 bucket 加總 amount_usd。';
COMMENT ON COLUMN public.coinglass_liq_map_level.snapshot_id IS '對應 coinglass_liq_map_snapshot.id';
COMMENT ON COLUMN public.coinglass_liq_map_level.exchange IS '交易所，如 Binance、OKX、Bybit';
COMMENT ON COLUMN public.coinglass_liq_map_level.instrument_id IS '交易對，如 BTCUSDT、BTC-USDT-SWAP';
COMMENT ON COLUMN public.coinglass_liq_map_level.bucket IS '價格區間，即 coinglass 圖表的橫軸；分析價位分佈請用此欄';
COMMENT ON COLUMN public.coinglass_liq_map_level.price IS '該檔位的精確價格；coinglass 頁面沒有使用';
COMMENT ON COLUMN public.coinglass_liq_map_level.amount_usd IS '該價位預估的清算金額（美元）';
COMMENT ON COLUMN public.coinglass_liq_map_level.leverage IS '槓桿倍數（10/25/50/100）；exchange 類型一律為 NULL';
COMMENT ON COLUMN public.coinglass_liq_map_level.tag IS 'h1/h2/h3，意義未知，coinglass 前端也未使用；分析時請忽略，不要推測含義。exchange 類型為 NULL';

-- 📋 coinglass_collector_run
COMMENT ON TABLE public.coinglass_collector_run IS
    'Coinglass 收集器的執行紀錄，每次執行 append 一列。分析前用來確認資料新鮮度：查每個 job 最近一筆的 started_at 與 status。job 對應：etf_flow、etf_spot_inflow、fear_greed（每 6 小時）、global_m2、liq_map（每 24 小時）、coin_flow（每 1 小時）。';
COMMENT ON COLUMN public.coinglass_collector_run.id IS '流水號';
COMMENT ON COLUMN public.coinglass_collector_run.job IS '任務名稱：etf_flow、etf_spot_inflow、fear_greed、global_m2、coin_flow、liq_map';
COMMENT ON COLUMN public.coinglass_collector_run.started_at IS '開始時間';
COMMENT ON COLUMN public.coinglass_collector_run.finished_at IS '結束時間';
COMMENT ON COLUMN public.coinglass_collector_run.status IS 'ok 或 error';
COMMENT ON COLUMN public.coinglass_collector_run.rows IS '本次寫入（含覆蓋更新）的列數';
COMMENT ON COLUMN public.coinglass_collector_run.error IS '失敗時的錯誤訊息';
COMMENT ON COLUMN public.coinglass_collector_run.login_error IS 'true 表示清算地圖的登入 token 失效，之後的清算地圖快照會缺漏，直到人工更新 token';
