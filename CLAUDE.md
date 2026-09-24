# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 專案性質

Spider-Hub 是一個**爬蟲逆向技術的範例集合**，不是一個可部署的服務。`cmd/spider-hub/main.go` 只印出說明文字，真正的程式碼都在 `platforms/` 底下，每個平台是一個獨立的 Go package，彼此不互相 import，也沒有共用的抽象層或介面。新增平台 = 新增一個資料夾，不需要修改任何既有檔案。

## 常用指令

```bash
# 建置檢查（單一平台）
cd platforms/pexels && go build -o /dev/null .

# 跑單一平台的全部測試
cd platforms/coinglass && go test -v

# 跑單一測試（測試會印大量 stdout，-v 是必要的）
cd platforms/amazon && go test -run TestAmazonProductDetail -v

# 全專案格式化 / 檢查
go fmt ./...
gofmt -l .
```

**測試都會打真實網路**：沒有 mock、沒有 fixture，`example_test.go` 直接請求線上 API。因此測試失敗常常是目標站台改版、被風控、或需要代理，而不是程式碼有 bug — 判斷前先看錯誤是 HTTP 狀態碼、解密失敗，還是解析欄位為空。`go test ./...` 會一次打所有平台的線上 API，通常不該這樣跑。

已知狀態：`platforms/pexels/` 尚未 gofmt（用空白縮排），`gofmt -l .` 會列出它。

## 平台目錄的兩種形態

沿用已刪除的 `SPIDER_HUB_RULES.md`（見 commit `e034969`）定下的規範，新平台先判斷屬於哪一類：

- **統一介面型**（coinglass）：所有 API 回傳同一種結構，主檔案叫 `spider.go`，型別叫 `Spider`、建構子 `NewSpider()`，對外只有一個通吃的方法（`GetData(apiURL)`）。
- **多樣介面型**（amazon、pexels、foresightnews）：各介面回傳結構差異大，主檔案以平台命名（`amazon.go`），型別叫 `<Platform>Spider`、建構子 `New<Platform>Spider()`，每個功能一個方法、各自回傳專屬結構。

固定檔案配置：主檔案、`types.go`（全部資料結構）、`constants.go`（URL／請求標頭／分頁上限／錯誤訊息字串）、`example_test.go`（唯一的測試入口，輔助函式也放這）、`README.md`（**每個平台只能有一個 README**）。可選：`utils.go`、`extractor.go`。不要另開測試目錄、不要拆出第二份 README、不要把設定散落在 constants.go 之外。

## 各平台的逆向要點

看程式碼前先知道每個平台卡在哪，改動時才不會破壞關鍵環節：

- **coinglass** — 雙層 AES-ECB + gzip。請求帶 `cache-ts-v2`（毫秒時間戳）與 `encryption: true`；回應的 `user` header 是一把加密過的動態金鑰，用「金鑰來源 base64 後取前 16 位」解開它（來源由回應標頭 `v` 決定，見下），再用解出的動態金鑰解 body 的 `data`，得到 gzip，解壓後才是 JSON。`GetData(url)` 是唯一的抓取原語，可打任何同協定端點；ETF 三支已包成 typed 方法（`GetETFFlow` 等）。
  - **找新端點的方法**：路徑猜不出來，去抓 `www.coinglass.com/<lang>/<page>` 的 Next.js chunks（`grep -oE '/_next/static/[^"]+\.js'` 下載後 grep `capi.coinglass.com/api/...`），裡面有完整路徑表。命名規律是 **BTC 不帶幣種前綴**（`/api/etf/flow`），其他幣才帶（`/api/etf/eth/flow`）。
  - 易混淆：`/api/etf/flow` 是每日淨流入流出，`/api/etf/history/chart?type=2` 是持倉存量/AUM。
  - **金鑰來源由回應標頭 `v` 決定**（2026-09-23 伺服端從 `v=0` 切到 `v=1`，曾導致全平台解密失效）。還原自前端 `mn()`：`v=0` → 請求送出的 `cache-ts-v2`；`v=1` → **API 路徑**（從 `/api` 起、不含 query）；`v=2` → 回應標頭 `time`；`v=55/66/77` → 三個固定字串（見 `constants.go` 的 `KeySourceV*`）。取 `base64(來源)[:16]` 當 AES key，其餘流程不變。`deriveKeySource()` 已實作全部分支，遇到未知 `v` 會明確報錯而非回傳亂碼。
    - 注意 `ev` 標頭（目前為 2）**不是**切換旗標，是誤導 —— 它不在 `Access-Control-Expose-Headers` 名單裡，瀏覽器根本讀不到。認 `v`。
    - **反混淆方法**（下次再改時直接重跑）：`app-*.js` 裡的 interceptor 用 javascript-obfuscator 的字串表混淆，解碼器 `Mn`（別名 `bn`）配字串陣列 `wn`，外加一個 checksum `191875` 的旋轉 IIFE。把這三塊原封不動抽進 node，就能把 `bn(341,"AQ3U")` 這類呼叫全部還原成字面量。金鑰推導在 `mn()`，AES 解密在 `xn()`。`gn`/`vn` 是另一組同構的解碼器。
    - 驗證訣竅：解對的話 `user` header 解密後是 gzip（magic `1f 8b`），解壓即得動態金鑰。用這個當 oracle 可離線掃金鑰，**不必反覆打線上 API**。
  - **熱力圖分四種，狀態不同**（皆 2026-09-23 實測）：首頁幣種熱力圖 `/api/futures/coins/heatmap?type=price|vol|oi` ✅ 可用（1612 幣，已包成 `GetCoinsHeatmap`，`type` 只吃這三個字串，傳數字會回 Server Error）；清算熱力圖 🔒 需登入；資金費率熱力圖 `/api/fundingRate/heatmap` 回 `null`；交易所熱力圖 `/api/derivative/exchange/heatmap` 的 `value` 恆為 0。後三者都是在解密正常時複測的真實回應。
  - **清算熱力圖**端點是 `/api/index/v2/liqHeatMap`（不是 `/api/coin/liq/heatmap`），參數 `merge/symbol/interval/limit/data`：`symbol` 用 `Binance_BTCUSDT` 格式、`interval` 是 `h24` 不是 `24h`、`limit` 預設 288。`data` 是 TOTP 簽章（secret `I65VU7K5ZQL7WB4E`、step 30，拼成 `"<unixsec>,<6位碼>"` 後以 AES-256-ECB／key `1f68efd73f8d4921acc0dead41dd39bc` 加密再 base64）。簽章已驗證會被接受，但接口另外要求登入（回 `{"code":"40000"}`，即前端彈登入框的碼），故未實作。
- **amazon** — 純 HTML 解析（goquery）+ 圖搜。圖搜流程是三步：先抓頁面裡的 `stylesnap` token → 下載圖片 → multipart 上傳搜尋；整串包在 `MaxRetries` 迴圈裡，任何一步失敗就整輪重來，因為 token 會失效。`extractor.go` 負責從 HTML 抽欄位，`utils.go` 負責價格／幣別／URL 驗證。
- **foresightnews** — 回應的 `data` 欄位是 base64 + zlib。注意 `code != 0` 並非錯誤（該站 `code` 為 1 也是正常回應），程式碼裡刻意註解掉了 code 檢查，不要「修好」它。用標準 `net/http`，不是 resty。
- **pexels** — 走 Next.js 的 `_next/data/<buildId>/...json` 端點，靠 `req` 的 `ImpersonateSafari()` 做 TLS/HTTP2 指紋偽裝過檢測。`NextDataBaseURL` 裡的 buildId 會隨對方部署變動，抓不到資料時先確認這個常數。目前 client 用 `req.DevMode()` 建立，會把完整請求／回應 dump 到 stdout。

三套 HTTP client 並存（resty、imroc/req、net/http）是刻意的：需要瀏覽器指紋偽裝時用 `req`，一般請求用 `resty`，簡單情境用標準庫。`go.mod` 裡多數依賴標為 `// indirect`，實際上是直接使用的，新增依賴後跑 `go mod tidy` 會修正。

## 慣例

- 錯誤與日誌訊息用中文並帶 emoji 前綴：`fmt.Errorf("❌ 請求失敗: %w", err)`、`fmt.Printf("✅ ...")`、`⚠️` 表警告。包裝錯誤用 `%w`。
- 對外方法第一個參數是 `context.Context`（coinglass 的 `GetData` 是早期例外）。
- 程式碼註解密度高、帶 emoji 標記，新增程式碼請沿用同樣風格。
- 分頁參數在方法內自行修正（`page < 1` 改成 1、`size` 超過 `MaxPageSize` 退回 `DefaultPageSize`），不回傳錯誤。
- 代理以 `SetProxy(proxyURL)` 方法或 `proxies map[string]string` 參數傳入，視平台而定。
