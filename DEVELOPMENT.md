# 🧑‍💻 開發流程

## 系統統計
- **API 端點**: 49 個 (100% 已實現)
- **資料表**: 5 張核心表
- **服務模組**: 11 個專業服務  
- **架構**: Go 1.23 + Gin + PostgreSQL + Bun ORM

## 常用指令

### 日常開發
```bash
make install        # 安裝依賴與 swag
make dev           # 生成文檔並啟動（推薦）
make build         # 編譯應用
make clean         # 清理構建文件
```

### 測試
```bash
./tests/test-all.sh              # 統一測試套件（推薦）
./tests/test-all.sh --type api   # API 功能測試
./tests/test-all.sh --type chat  # 對話功能測試
make test                       # Go 單元測試
```

### 資料庫管理
```bash
make db-setup      # 初始化 + 執行遷移
make migrate       # 執行待處理遷移
make migrate-status # 查看遷移狀態
make fixtures      # 載入種子資料
make migrate-reset # 重置所有遷移（需確認）
```

### 快速工作流
```bash
make fresh-start   # 完整重建（清理+安裝+資料庫+種子資料）
make quick-setup   # 快速設置（資料庫+種子資料）
```

## 代碼規範

- **語言**: Go 1.23+，提交前保持 go fmt 乾淨
- **命名**: 套件小寫無底線，檔案snake_case.go
- **識別符**: 匯出UpperCamelCase，區域變數lowerCamelCase  
- **JSON**: snake_case標籤（`json:"should_use_grok"`）
- **架構**: 函式小而專注，services/內用建構子注入依賴

## CLI 工具

### 遷移管理
| 功能 | Make命令 | 直接命令 |
|------|----------|----------|
| 初始化遷移表 | `make db-init` | `go run cmd/bun/main.go db init` |
| 執行遷移 | `make migrate` | `go run cmd/bun/main.go db migrate` |
| 回滾遷移 | `make migrate-down` | `go run cmd/bun/main.go db rollback` |
| 查看狀態 | `make migrate-status` | `go run cmd/bun/main.go db status` |
| 重置遷移 | `make migrate-reset` | `go run cmd/bun/main.go db reset` |
| 載入數據 | `make fixtures` | `go run cmd/bun/main.go db fixtures` |

### 文件結構
```
cmd/bun/
├── main.go              # CLI 工具入口
├── migrations/          # Go 遷移文件
└── fixtures/           # 種子數據
    └── fixtures.yml
```

## 測試系統

### 測試執行
```bash
./tests/test-all.sh              # 所有測試
./tests/test-all.sh --type api   # API測試  
./tests/test-all.sh --type chat  # 對話測試
./tests/test-all.sh --csv        # 生成CSV報告
```

### 覆蓋範圍
- **系統健康**: 服務器連接、API版本、監控端點
- **API功能**: 用戶認證、角色系統、情感系統、搜索、TTS
- **對話功能**: 會話管理、多場景對話、AI引擎、情感追蹤
- **NSFW分級**: 5級分類、準確率驗證、引擎選擇

## API 架構

### 文檔與端點
- **Swagger UI**: `/swagger/index.html`（即時生成）
- **進度追蹤**: `API_PROGRESS.md`
- **總端點**: 49個（系統7個+認證4個+用戶4個+角色8個+對話9個+情感3個+搜索2個+TTS2個+管理10個）

### 路由結構
- **定義**: `routes/routes.go`
- **處理器**: `handlers/` 各模組
- **中間件**: `middleware/` 認證授權

## 推薦工作流

```bash
# 日常開發
make dev                    # 文檔生成 + 啟動

# 資料庫設置  
make fresh-start           # 完整重建
make quick-setup           # 快速設置

# 測試驗證
./tests/test-all.sh        # 統一測試

# 新功能開發
make create-migration NAME=feature_name  # 創建遷移
make migrate-status        # 檢查狀態
```

