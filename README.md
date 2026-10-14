# 🕷️ Spider-Hub

> 一個專注於各平台爬蟲逆向技術的 Go 語言專案集合，外加一個把 Coinglass 資料寫入 PostgreSQL 的收集器

[![Go Version](https://img.shields.io/badge/Go-1.24+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-Apache%202.0-green.svg)](LICENSE)

> **出處**：本專案 fork 自 [xieburoucoco/spider-hub](https://github.com/xieburoucoco/spider-hub)（Apache 2.0）。在原專案基礎上新增了 Coinglass 的恐懼貪婪指數、全球 M2、現貨／合約資金流入流出、清算地圖等接口，以及定期寫入 PostgreSQL 的收集器 [`collectors/coinglass`](collectors/coinglass/README.md)。原專案的著作權與授權條款維持不變，見 [LICENSE](LICENSE)。

## 📖 專案簡介

Spider-Hub 收集各網站的爬蟲逆向實作，每個平台是一個獨立的 Go package，記錄該站的加密、簽章、指紋檢測等關卡，以及對應的破解方式，供學習與參考。

`platforms/` 底下都是範例程式碼；唯一可部署的程式是 `collectors/coinglass`。

## ✨ 功能特性

- 🔍 **多平台支援**：涵蓋電商、圖庫、新聞資訊、加密貨幣數據等平台
- 🔐 **逆向紀錄**：每個平台的 README 記錄加密協定、簽章演算法與找端點的方法
- 🗄️ **資料收集器**：Coinglass 資料定期入庫，排程、增量回補、登入保護都已內建
- 🤖 **給 AI agent 用的資料說明**：資料表與欄位說明寫在資料庫註解裡，另附 [DATA.md](collectors/coinglass/DATA.md) 使用指南

## 🏗️ 專案結構

```
spider-hub/
├── README.md                 # 專案說明文件
├── LICENSE                   # 開源協議（Apache 2.0）
├── CLAUDE.md                 # 給 AI 程式助手的開發說明
├── go.mod / go.sum           # Go 模組檔案
├── cmd/
│   └── spider-hub/           # 主程式入口（只印說明文字）
├── platforms/                # 各平台爬蟲實作（彼此獨立、不互相 import）
│   ├── amazon/               # Amazon 商品詳情與圖搜
│   ├── coinglass/            # Coinglass 加密貨幣數據
│   ├── foresightnews/        # Foresight News 快訊
│   └── pexels/               # Pexels 影片搜尋
└── collectors/
    └── coinglass/            # Coinglass → PostgreSQL 常駐收集器
```

## 🚀 快速開始

### 環境需求

- Go 1.24+
- Git
- PostgreSQL（只有收集器需要）

### 安裝步驟

1. **複製專案**
```bash
git clone git@github.com:yushenlin118/spider-hub.git
cd spider-hub
```

2. **安裝依賴**
```bash
go mod download
```

> Go module 路徑仍是 `github.com/xieburoucoco/spider-hub`，在本 repo 內建置不受影響。

## 📚 使用指南

### 平台使用

每個平台都有獨立的實作與文件，請進入對應目錄查看：

```bash
# 查看 Coinglass 平台說明
cd platforms/coinglass
cat README.md

# 執行單一測試（測試會印大量輸出，-v 是必要的）
go test -run TestFearGreedIndex -v
```

> ⚠️ 測試都會直接請求線上 API，沒有 mock。失敗時常是目標站台改版、被風控或需要代理，不一定是程式碼問題。不建議執行 `go test ./...`，那會一次打所有平台的線上 API。

### 平台列表

| 平台 | 內容 | 逆向重點 | 文件 |
|---|---|---|---|
| **Coinglass** | ETF 流入流出、熱力圖、清算地圖、恐懼貪婪、全球 M2、資金流入流出 | 雙層 AES-ECB + gzip，金鑰來源依回應標頭 `v` 切換；清算接口需 TOTP 簽章與登入 | [查看詳情](platforms/coinglass/README.md) |
| **Amazon** | 商品詳情、以圖搜圖 | HTML 解析；圖搜需先取 `stylesnap` token | [查看詳情](platforms/amazon/README.md) |
| **Foresight News** | 新聞快訊 | 回應 `data` 為 base64 + zlib | [查看詳情](platforms/foresightnews/README.md) |
| **Pexels** | 影片搜尋 | Next.js data 端點 + Safari TLS 指紋偽裝 | [查看詳情](platforms/pexels/README.md) |

### Coinglass 收集器

定期抓取 Coinglass 資料寫入 PostgreSQL，只會在 `public` schema 新增 `coinglass_*` 資料表，不會碰其他資料表：

```bash
# 所有任務立刻跑一次
DATABASE_URL='postgres://...' go run ./collectors/coinglass -once

# 常駐模式（內建排程，重啟不會重抓）
DATABASE_URL='postgres://...' go run ./collectors/coinglass
```

任務頻率、環境變數、Docker 部署與維運方式見 [collectors/coinglass/README.md](collectors/coinglass/README.md)；資料怎麼查詢見 [DATA.md](collectors/coinglass/DATA.md)。

## 🔧 開發指南

### 新增平台

1. 在 `platforms/` 下建立新的平台目錄，不需要修改任何既有檔案
2. 依介面形態選擇命名：所有 API 結構相同用 `spider.go` / `Spider`，結構各異用 `<platform>.go` / `<Platform>Spider`
3. 固定檔案：主檔案、`types.go`、`constants.go`、`example_test.go`、`README.md`（每個平台只有一份）
4. 更新本文件的平台列表

### 程式碼規範

- 遵循 Go 官方程式碼規範
- 使用 `gofmt` 格式化程式碼
- 錯誤與日誌訊息用中文並帶 emoji 前綴（`❌`、`✅`、`⚠️`）
- 對外方法第一個參數是 `context.Context`

## 📊 技術棧

- **語言**：Go 1.24+
- **HTTP 客戶端**：resty（一般請求）、imroc/req（瀏覽器指紋偽裝）、net/http（簡單情境）
- **資料解析**：goquery（HTML）、encoding/json
- **資料儲存**：PostgreSQL（pgx，僅收集器使用）

## 📄 開源協議

本專案沿用原專案的 [Apache 2.0](LICENSE) 開源協議。

## 🔗 相關連結

- [本專案](https://github.com/yushenlin118/spider-hub)
- [原專案 xieburoucoco/spider-hub](https://github.com/xieburoucoco/spider-hub)

## ⚠️ 免責聲明

本專案僅供學習與研究使用，請遵守相關網站的使用條款與 robots.txt 協議。使用者須自行承擔使用風險，作者不承擔任何法律責任。
