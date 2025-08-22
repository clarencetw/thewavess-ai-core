# 🔄 CLI 工具遷移指南

## 📝 概述
已將分散的 CLI 工具整合為符合 Bun 官方最佳實踐的統一工具。

## 🔄 命令對照表

| 功能 | 舊命令 | 新命令 | Make 命令 |
|------|--------|--------|-----------|
| 🏗️ 初始化遷移表 | - | `go run cmd/bun/main.go db init` | `make db-init` |
| ⬆️ 執行遷移 | `go run cmd/migrate/main.go -cmd=up` | `go run cmd/bun/main.go db migrate` | `make migrate` |
| ⬇️ 回滾遷移 | `go run cmd/migrate/main.go -cmd=down` | `go run cmd/bun/main.go db rollback` | `make migrate-down` |
| 📊 遷移狀態 | `go run cmd/migrate/main.go -cmd=status` | `go run cmd/bun/main.go db status` | `make migrate-status` |
| 🔄 重置遷移 | - | `go run cmd/bun/main.go db reset` | `make migrate-reset` |
| 🌱 載入 Fixtures | - | `go run cmd/bun/main.go db fixtures` | `make fixtures` |
| 🔄 重建並載入 Fixtures | - | `go run cmd/bun/main.go db fixtures --recreate` | `make fixtures-recreate` |
| ➕ 創建遷移 | - | `go run cmd/bun/main.go create-migration <name>` | `make create-migration NAME=<name>` |

## 🚀 快速開始

```bash
# 1. 完整資料庫設置
make db-setup

# 2. 載入 fixtures 數據  
make fixtures

# 3. 檢查狀態
make migrate-status

# 或者使用一鍵設置（推薦）
make fresh-start    # 清理+安裝+資料庫設置+fixtures
make quick-setup    # 資料庫設置+fixtures（不清理）
```

## 📁 文件結構變化

```
cmd/
└── bun/
    ├── main.go           # ✅ 統一 CLI 工具
    ├── migrations/       # ✅ Go-based 遷移文件
    │   ├── main.go
    │   ├── 20250815000001_users.go
    │   ├── 20250815000002_characters.go
    │   └── ...
    └── fixtures/
        └── fixtures.yml  # ✅ 種子數據配置
```

## 🆕 新增的 Makefile 命令

除了基本遷移命令外，還新增了以下便利命令：

```bash
# 🏗️ 資料庫相關
make db-setup          # 初始化 + 遷移
make fresh-start       # 完整重建（清理+安裝+資料庫+數據）
make quick-setup       # 快速設置（資料庫+數據）

# 📊 開發相關
make dev              # 生成文檔 + 開發模式運行
make docs             # 生成 Swagger 文檔
make test-api         # 後台運行 + API 測試

# 🔍 監控相關  
make check            # 檢查服務狀態
make run-bg           # 後台運行服務
make stop-bg          # 停止後台服務
```


## 💡 推薦做法

- ✅ 使用 Make 命令進行日常操作
- ✅ 使用 `make create-migration NAME=xxx` 創建新遷移
- ✅ 開發時使用 `make dev` 一次性完成文檔生成和服務啟動
- ✅ 部署前使用 `make fresh-start` 確保完整重建
- ✅ 快速重建使用 `make quick-setup` 跳過清理步驟
- ✅ API 測試使用 `make test-api` 自動後台運行並測試

## ⚙️ CLI 工具詳細用法

### 資料庫管理
```bash
# 直接使用 CLI 工具（不推薦，建議使用 Make 命令）
go run cmd/bun/main.go db init           # 初始化遷移表
go run cmd/bun/main.go db migrate        # 執行遷移
go run cmd/bun/main.go db rollback       # 回滾遷移
go run cmd/bun/main.go db status         # 查看狀態
go run cmd/bun/main.go db reset          # 重置所有遷移
```

### Fixtures 管理
```bash
go run cmd/bun/main.go db fixtures              # 載入 fixtures
go run cmd/bun/main.go db fixtures --recreate   # 重建表格並載入 fixtures
```

### 遷移文件創建
```bash
go run cmd/bun/main.go create-migration add_user_table
# 將創建: cmd/bun/migrations/20250822000001_add_user_table.go
```