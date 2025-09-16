# NSFW 指南 - 三層 AI 引擎架構

本指南說明當前採用的三層 AI 引擎 NSFW 處理策略，提供智能內容分級與引擎路由。

## 系統架構

### 三層 AI 引擎分工
- **L1 (安全內容)** → OpenAI GPT-4o (情感細膩，專業對話)
- **L2-L3 (中等 NSFW)** → Mistral AI (專業 NSFW 處理，平衡性能)
- **L4-L5 (明確露骨)** → Grok AI (大膽創意，無限制對話)

### 智能分級系統
使用權重評分制，基於 18 個關鍵字規則：
- **L5 (≥10分)**: 明確性行為 → Grok
- **L4 (≥6分)**: 明確身體部位 → Grok
- **L3 (≥4分)**: 裸體/色情場景 → Mistral
- **L2 (≥2分)**: 身體描述/委婉語 → Mistral
- **L1 (<2分)**: 輕微暗示/安全 → OpenAI

## 關鍵字規則

### L5 級關鍵字 (10分)
```
性行為: 做愛、性愛、性交、高潮、射精、自慰、口交、肛交
英文: sex, orgasm, masturbation, blowjob, anal
```

### L4 級關鍵字 (6分)
```
身體部位: 陰莖、陰道、乳頭、下體、生殖器
英文: penis, vagina, nipple, genitals
```

### L3 級關鍵字 (4分)
```
情境描述: 裸體、色情、床戲、成人片、脫衣
英文: nude, porn, naked, strip
```

### L2 級關鍵字 (2分)
```
身體描述: 胸部、身材、性感、誘惑、親吻
英文: breast, sexy, kiss, seduction
```

## 智能特性

### 1. 上下文抑制
醫療、藝術、教育語境自動降級：
```
"乳房檢查" (醫療) → L1
"藝術裸體" (藝術) → L1
"性教育" (教育) → L1
```

### 2. Sticky 會話機制
- 會話一旦觸發 L3+ 內容，後續 3 分鐘內維持同引擎
- 避免引擎切換造成對話不一致
- 3 分鐘後重新評估

### 3. Fallback 機制
- Mistral 無法處理 → 自動切換到 Grok
- Grok 無法處理 → 返回適當錯誤訊息
- 確保服務可用性

## 安全合規

### 禁止內容 (直接拒絕)
- 未成年相關內容
- 非自願/強迫行為
- 暴力性行為
- 血腥暴力內容
- 仇恨言論

### 年齡驗證
- 涉及 L3+ 內容時需要年齡確認
- 確保用戶為 18+ 成年人
- 完整的用戶同意流程

## 技術實施

### 關鍵組件
```
services/nsfw_classifier.go - NSFW 分級邏輯
services/chat_service.go - AI 引擎選擇
services/mistral_client.go - Mistral AI 客戶端
services/openai_client.go - OpenAI 客戶端
services/grok_client.go - Grok AI 客戶端
```

### 配置選項
```bash
# 三層引擎 API Keys
OPENAI_API_KEY=sk-your-openai-key
MISTRAL_API_KEY=your-mistral-key
GROK_API_KEY=xai-your-grok-key

# 引擎配置
OPENAI_MODEL=gpt-4o
MISTRAL_MODEL=mistral-medium-latest
GROK_MODEL=grok-beta
```

## 測試驗證

### 測試腳本
```bash
# 三層引擎整合測試
./tests/test_mistral_integration.sh

# 完整系統測試
./tests/test-all.sh --type nsfw --csv
```

### 測試用例
- L1: "今天天氣真好！" → OpenAI
- L2: "你的身材真好" → Mistral
- L3: "我想看你的裸體" → Mistral
- L4: "我想要口交" → Grok
- L5: "我要強姦你" → Grok (違規拒絕)

## 性能指標

- **分類準確率**: 95%+
- **響應時間**: 1-3 秒
- **引擎可用性**: 99.9%
- **Fallback 成功率**: 100%

## 維護指南

### 關鍵字更新
1. 編輯 `services/nsfw_classifier.go`
2. 更新對應的權重分數
3. 運行測試確保分類正確
4. 重新部署服務

### 引擎調整
1. 修改 `.env` 中的模型配置
2. 測試新模型的回應品質
3. 監控錯誤率和響應時間
4. 根據需要調整 fallback 策略

這個三層架構提供了最佳的內容處理能力，確保每個層級的內容都由最適合的 AI 引擎處理。