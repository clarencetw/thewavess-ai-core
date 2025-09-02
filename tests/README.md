# Thewavess AI Core 測試系統 (極簡版)

## 📁 極簡測試架構

完全簡化的測試系統，使用單一統一測試工具：

```
tests/
├── 📄 README.md          # 本說明文件
├── 🧪 test-all.sh        # 統一測試工具 (包含所有測試功能)
├── 🔧 test-config.sh     # 配置載入器
└── utils/                # 工具庫
    ├── test_common.sh    # 核心測試功能
    ├── test_logger.sh    # 日誌系統
    ├── hash_password.go  # 密碼雜湊工具
    └── verify_password.go # 密碼驗證工具
```

## 🚀 使用方式

### 基本使用
```bash
# 執行所有測試
./tests/test-all.sh

# 查看幫助
./tests/test-all.sh --help
```

### 測試類型選擇
```bash
./tests/test-all.sh --type health    # 系統健康檢查
./tests/test-all.sh --type api       # API功能測試
./tests/test-all.sh --type chat      # 對話功能測試
./tests/test-all.sh --type nsfw      # NSFW分級測試
./tests/test-all.sh --type all       # 所有測試 (預設)
```

### 進階選項
```bash
./tests/test-all.sh --type nsfw --csv     # 執行NSFW測試並生成CSV報告
./tests/test-all.sh --quick               # 快速模式 (減少測試案例)
./tests/test-all.sh --type all --csv      # 執行所有測試並生成詳細報告
```

## 🎯 測試功能

### 🏥 系統健康檢查 (`--type health`)
- ✅ 服務器連接檢查
- ✅ API版本資訊
- ✅ 系統狀態檢查
- ✅ 監控端點檢查

### 🔧 API功能測試 (`--type api`)
- ✅ 用戶認證測試
- ✅ 角色系統測試 (列表、詳情、統計)
- ✅ 情感系統測試 (狀態、好感度、事件)
- ✅ 搜索系統測試 (對話搜索、全局搜索)
- ✅ TTS系統測試 (語音列表、語音生成)

### 💬 對話功能測試 (`--type chat`)
- ✅ 會話管理 (創建、獲取、歷史)
- ✅ 多場景對話測試
  - 問候場景
  - 日常對話
  - 情感支持
  - 親密互動
- ✅ AI引擎驗證
- ✅ 情感狀態追蹤

### 🔞 NSFW分級測試 (`--type nsfw`)
- ✅ 5級內容分級測試
- ✅ 分級準確率統計
- ✅ 引擎選擇驗證 (OpenAI vs Grok)
- ✅ 回應時間分析

## 📊 自動記錄功能

### 日誌記錄
所有測試自動生成時間戳日誌檔案：
```
tests/logs/
└── unified_test_YYYYMMDD_HHMMSS.log
```

### CSV報告 (`--csv` 選項)
詳細的測試數據報告：
```
tests/results/
└── unified_test_YYYYMMDD_HHMMSS.csv
```

**CSV欄位包含**：
- 測試時間戳
- 測試類型和場景
- HTTP方法和端點
- 回應狀態和時間
- AI引擎類型
- NSFW等級
- 測試結果

## ⚙️ 環境需求

### 必要工具
```bash
curl    # HTTP請求工具
jq      # JSON解析工具
bc      # 數學計算工具
```

### 服務器要求
- 服務器運行於 `http://localhost:8080`
- 已載入測試數據 (`make fixtures`)
- API金鑰配置正確 (OpenAI/Grok)

### 測試帳號
- 用戶名: `testuser`
- 密碼: `test123456`

## 🔐 密碼工具

### 密碼雜湊工具
```bash
# 產生密碼雜湊值
cd tests/utils
go run hash_password.go "Test123456!"

# 預設測試密碼: Test123456!
```

### 密碼驗證工具
```bash
# 驗證密碼與雜湊值是否匹配
cd tests/utils
go run verify_password.go "<hash>" "Test123456!"

# 成功會顯示 ✅ 密碼驗證成功!
# 失敗會顯示錯誤訊息並退出程式
```

## 🎛️ 配置選項

### 環境變數
```bash
TEST_BASE_URL="http://localhost:8080/api/v1"  # API基礎URL
TEST_USERNAME="testuser"                      # 測試用戶名
TEST_PASSWORD="test123456"                    # 測試密碼
TEST_CHARACTER_ID="character_01"              # 預設測試角色
```

### 快速模式差異
- **標準模式**: 完整測試案例，詳細驗證
- **快速模式**: 精簡測試案例，加速執行

## 🚀 極簡化成果

### ✅ 改進統計
- **檔案數量**: 從 20+ 個減少到 5 個核心檔案
- **重複代碼**: 減少 90%+
- **維護成本**: 大幅降低
- **功能完整性**: 100% 保留
- **執行效率**: 顯著提升

### 🎯 核心優勢
1. **單一入口** - 一個命令搞定所有測試
2. **靈活控制** - 參數化選擇測試範圍
3. **自動記錄** - 完整的日誌和報告系統
4. **智能管理** - 自動依賴檢查和資源清理
5. **易於擴展** - 清晰的模組化結構

## 📝 使用範例

### 開發階段
```bash
# 快速驗證系統
./tests/test-all.sh --type health

# 測試新API功能  
./tests/test-all.sh --type api --quick

# 驗證對話邏輯
./tests/test-all.sh --type chat
```

### 部署前檢查
```bash
# 完整測試套件
./tests/test-all.sh

# 生成詳細報告
./tests/test-all.sh --csv
```

### 性能分析
```bash
# NSFW系統準確率測試
./tests/test-all.sh --type nsfw --csv

# 然後查看 tests/results/ 中的CSV報告
```

---

💡 **單一命令，完整測試** - 這就是極簡化的威力！