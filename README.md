# Thewavess AI Core

> 🤖 專為女性用戶設計的智能 AI 聊天後端服務

Thewavess AI Core 是一個基於 Golang 開發的企業級 AI 聊天後端，整合 OpenAI 和 Grok 引擎，提供智能對話、情感陪伴、角色互動等核心功能。

## ✨ 核心特色
- **🎭 角色系統**: 霸道總裁、溫柔醫生等個性化 AI 角色
- **💕 情感互動**: 好感度、關係發展、情感記憶
- **🔞 內容分級**: 5級智能內容分析，支援成人對話
- **🎨 場景描述**: 動態場景生成，增強沉浸感
- **📱 多平台**: RESTful API，支援 Web/App 集成

## 🚀 快速開始

### 最小化部署
```bash
# 1. 複製項目
git clone https://github.com/clarencetw/thewavess-ai-core.git
cd thewavess-ai-core

# 2. 設定環境變數
cp .env.example .env
# 編輯 .env，填入 OPENAI_API_KEY

# 3. 啟動服務
go run main.go
```

### 驗證部署
- 🌐 **Web 介面**: http://localhost:8080
- 📚 **API 文檔**: http://localhost:8080/swagger/index.html
- ❤️ **健康檢查**: http://localhost:8080/health

## 🛠️ 技術架構
- **後端**: Go + Gin + 結構化日誌
- **AI 引擎**: OpenAI GPT-4o + Grok (NSFW)
- **內容分析**: 智能分級系統 (Level 1-5)
- **部署**: Docker 容器化支援

## 📋 功能狀態

### ✅ 已實現核心功能
- **智能對話系統**: OpenAI GPT-4o 整合完成
- **5級內容分級**: 支援從日常對話到成人內容
- **角色個性化**: 陸寒淵、沈言墨角色系統
- **場景描述**: 動態生成沉浸式場景
- **結構化日誌**: 完整的監控和調試支援
- **環境配置**: godotenv 自動載入

### 🔄 開發中功能
- **Grok API 整合**: Level 5 內容處理
- **數據庫持久化**: PostgreSQL + Redis 整合
- **記憶系統**: 長期記憶和向量檢索
- **TTS 語音**: 語音合成功能

### 📊 API 端點進度
目前已實現 **22/118 個端點**，核心對話功能完全可用。

## 📚 文檔導航

| 文檔 | 用途 | 適用對象 |
|------|------|----------|
| **[README.md](./README.md)** | 項目概覽、快速開始 | 所有用戶 |
| **[API.md](./API.md)** | 完整 API 參考文檔 | 開發者 |
| **[API_PROGRESS.md](./API_PROGRESS.md)** | API 開發進度追蹤 | 開發者、PM |
| **[SPEC.md](./SPEC.md)** | 產品規格、技術架構 | 架構師、PM |
| **[NSFW_GUIDE.md](./NSFW_GUIDE.md)** | NSFW 功能詳細說明 | 開發者、合規 |
| **[DEPLOYMENT.md](./DEPLOYMENT.md)** | 部署和運維指南 | DevOps、運維 |
| **[CLAUDE.md](./CLAUDE.md)** | Claude Memory Guide | AI Assistant |

### 🎯 快速導航
- **🚀 立即開始**: 看 [快速開始](#🚀-快速開始)
- **🔌 API 集成**: 看 [API.md](./API.md)
- **📊 開發進度**: 看 [API_PROGRESS.md](./API_PROGRESS.md)
- **🔞 成人內容**: 看 [NSFW_GUIDE.md](./NSFW_GUIDE.md)
- **☁️ 部署上線**: 看 [DEPLOYMENT.md](./DEPLOYMENT.md)
- **🏗️ 系統設計**: 看 [SPEC.md](./SPEC.md)

## 🔞 NSFW 內容支援

Thewavess AI Core 實現了業界領先的 **5級智能內容分級系統**，支援從日常對話到明確成人內容（包含性器官描述）的完整處理。

### 快速了解
- ✅ **Level 1-4**: OpenAI 處理，包含成人內容
- ✅ **Level 5**: Grok 處理極度明確內容  
- ✅ **智能檢測**: 自動分析用戶意圖並適配
- ⚠️ **18+ 限制**: 僅限成年用戶使用

詳細說明請參考 **[NSFW_GUIDE.md](./NSFW_GUIDE.md)**

## ⚖️ 合規聲明
- 🔞 面向 18 歲以上成年用戶
- 🚫 嚴禁未成年人相關內容
- 📜 用戶需遵守當地法律法規
- 🔒 敏感內容加密存儲

## 🤝 參與貢獻

### 開發者
1. Fork 此專案
2. 閱讀 [SPEC.md](./SPEC.md) 了解架構
3. 參考 [DEPLOYMENT.md](./DEPLOYMENT.md) 設置開發環境
4. 提交 Pull Request

### 問題回報
- 🐛 **Bug 回報**: 使用 GitHub Issues
- 💡 **功能建議**: 歡迎在 Issues 中討論
- 📚 **文檔改進**: 直接提交 PR

## 📄 授權信息
- **版權所有**: © 2024 clarencetw
- **授權類型**: 專有軟體，保留所有權利
- **使用限制**: 僅限授權用戶使用

---

<div align="center">

**🚀 讓 AI 對話更加智能和貼心 🚀**

[GitHub](https://github.com/clarencetw/thewavess-ai-core) • [文檔](./API.md) • [部署指南](./DEPLOYMENT.md)

</div>
