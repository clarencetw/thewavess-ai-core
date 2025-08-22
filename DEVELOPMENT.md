# 🧑‍💻 開發流程（Development）

—

## 常用指令（Makefile）

開發與執行：
```bash
make install   # 安裝依賴與 swag
make dev       # 生成 Swagger 並啟動
make run       # 僅啟動服務
make build     # 編譯到 bin/thewavess-ai-core
```

文件與測試：
```bash
make docs         # 生成 Swagger
make test         # go test -v ./...
make test-api     # 後台啟動並執行 test_api.sh
```

資料庫（PostgreSQL + Bun）：
```bash
make db-setup         # 初始化遷移表 + 遷移
make migrate          # 執行遷移
make migrate-status   # 查看狀態
make migrate-down     # 回滾一次
make seed             # 填充種子資料
```

Docker：
```bash
make docker-build
make docker-run
```

—

## 代碼規範

- 語言：Go 1.23+；提交前保持 go fmt 乾淨
- 套件命名：小寫、無底線；檔案以功能命名（snake_case.go）
- 匯出識別：UpperCamelCase；區域變數：lowerCamelCase
- JSON tag：snake_case（例：json:"should_use_grok"）
- 函式小而專注；偏好在 services/ 內用建構子注入依賴

—

## 測試

- 單元測試：與程式碼同層級，命名 *_test.go
- 表格驅動為佳；以 make test 執行所有測試
- 端點煙霧測試：make test-api（啟動服務後執行 test_api.sh）
- 模擬外部 API；勿在測試中呼叫真實雲端

—

## 路由與 Swagger

- 路由入口：routes/routes.go
- Swagger UI：/swagger/index.html（由 swag 根據 handler 註解產生）

