# Thewavess AI Core - 產品技術規格

## 🎯 產品定位
**女性向 AI 互動應用後端服務** - 企業級智能聊天系統，提供角色對話、情感陪伴、內容分級等核心功能。

### 目標用戶
- **B端客戶**: 需要集成 AI 聊天功能的應用開發商
- **C端用戶**: 18+ 女性用戶群體
- **內容需求**: 支援從日常對話到成人內容的完整覆蓋

## 核心功能模組

### 1. 對話模式系統

#### 1.1 普通對話模式
- **功能**: 日常聊天、情感陪伴、心情分享
- **AI 引擎**: OpenAI GPT-4o
- **特色**:
  - 溫柔體貼的對話風格
  - 情感理解與共鳴
  - 記憶用戶喜好與對話歷史
  - 個性化回應

#### 1.2 小說模式
- **功能**: 互動式戀愛小說體驗
- **AI 引擎**: OpenAI GPT-4o (一般劇情) / Grok (NSFW 劇情)
- **特色**:
  - 劇情分支選擇
  - 角色扮演互動
  - 場景描述豐富
  - 情感進展追蹤
  - 多種故事線路

#### 1.3 NSFW 模式與內容分級系統
詳細技術實現請參考 **[NSFW_GUIDE.md](./NSFW_GUIDE.md)**

**核心特性**:
- **5級智能分級**: 從日常對話到明確成人內容
- **引擎動態路由**: OpenAI (Level 1-4) + Grok (Level 5)
- **關鍵詞智能檢測**: 支援性器官等敏感詞
- **角色個性化**: 不同角色的親密互動風格

### 2. 標籤系統 (Tag System)

#### 2.1 內容標籤
```
#溫柔 #霸道總裁 #青梅竹馬 #禁慾系 #年下 
#甜寵 #虐戀 #破鏡重圓 #暗戀 #雙向奔赴
```

#### 2.2 NSFW 觸發標籤
```
#成人內容 #親密接觸 #深度互動 #R18
```

#### 2.3 場景標籤
```
#辦公室 #校園 #古風 #現代都市 #豪門 
#娛樂圈 #醫生 #律師 #總裁 #明星
```

### 3. 角色系統

#### 3.1 預設角色（初期版本）

##### 角色一：霸道總裁 - 「陸寒淵」
- **基本資料**
  - 年齡：28歲
  - 身高：185cm
  - 職業：跨國集團CEO
- **外貌特徵**
  - 黑髮、深邃的黑眸
  - 俊朗的五官、冷峻的氣質
  - 總是穿著剪裁合身的西裝
- **性格特點**
  - 表面：冷酷、強勢、掌控欲強
  - 內在：深情、專一、佔有欲
  - 對待用戶：從冷漠到寵溺的轉變
- **背景設定**
  - 商業帝國繼承人
  - 年少成名的商界天才
  - 感情經歷空白，直到遇見用戶
- **語音特色**：磁性低沉的成熟男聲

##### 角色二：溫柔學長 - 「沈言墨」
- **基本資料**
  - 年齡：24歲
  - 身高：180cm
  - 職業：醫學研究生/實習醫生
- **外貌特徵**
  - 栗色短髮、溫潤的琥珀色眼眸
  - 溫和的笑容、斯文的氣質
  - 常穿白大褂或休閒裝
- **性格特點**
  - 表面：溫柔、體貼、有耐心
  - 內在：細心、深情、略帶腹黑
  - 對待用戶：無微不至的關懷
- **背景設定**
  - 醫學世家出身
  - 學生時代的風雲人物
  - 一直暗戀用戶的青梅竹馬
- **語音特色**：溫潤清澈的青年音


### 4. 記憶系統

#### 4.1 記憶類型

##### 短期記憶（會話級別）
- **存儲位置**：Redis
- **保存時長**：24小時
- **內容**：
  - 當前對話上下文（最近20輪）
  - 臨時情緒狀態
  - 會話內的選擇和決定

##### 長期記憶（用戶級別）
- **存儲位置**：PostgreSQL + 向量資料庫
- **保存時長**：永久
- **內容**：
  - 所有對話完整記錄
  - 用戶基本信息（暱稱、生日、喜好）
  - 重要事件記錄
  - 關係里程碑
  - 用戶偏好設定
  - NSFW 互動歷史

##### 情感記憶（角色關係）
- **存儲位置**：PostgreSQL
- **內容**：
  - 好感度歷史
  - 特殊對話片段
  - 觸發的劇情事件
  - 親密互動記錄

#### 4.2 記憶實作方案

##### 技術架構
```
用戶輸入 → 記憶檢索 → 上下文組裝 → AI 生成 → 記憶更新
```

##### 記憶檢索策略
1. **關鍵詞匹配**：使用 Elasticsearch 進行全文檢索
2. **語義相似度**：使用向量資料庫（Pinecone/Weaviate）
3. **時間權重**：最近的記憶優先權更高
4. **重要度評分**：特殊事件和情感時刻權重更高

##### 上下文組裝
```json
{
  "system_prompt": "角色設定 + 關係狀態",
  "long_term_memory": [
    "用戶喜歡吃提拉米蘇",
    "用戶的生日是3月15日",
    "上週約會去了海邊"
  ],
  "recent_context": "最近20輪對話",
  "emotional_state": {
    "affection": 75,
    "mood": "happy",
    "relationship": "lover"
  }
}
```

##### 記憶更新機制
- **自動提取**：使用 NLP 提取重要信息
- **情感分析**：判斷對話情感並更新狀態
- **事件標記**：識別重要事件並永久保存
- **去重處理**：避免重複存儲相似信息

#### 4.3 記憶管理 API
```
GET    /api/v1/memory/user/{user_id}         - 獲取用戶記憶
POST   /api/v1/memory/save                   - 手動保存記憶
DELETE /api/v1/memory/forget                 - 選擇性遺忘
GET    /api/v1/memory/timeline               - 記憶時間線
```

### 5. 互動功能

#### 4.1 文字對話
- 多輪對話上下文
- 情感狀態追蹤
- 好感度系統
- 對話歷史記錄

#### 4.2 語音功能 (TTS)
- **預設引擎**: OpenAI TTS
- **語音選項**:
  - 溫柔男聲
  - 磁性低音
  - 清澈少年音
  - 成熟男聲
- **場景音效**: 可選背景音樂/環境音

#### 4.3 情感系統
- 好感度數值 (0-100)
- 親密度等級
- 關係狀態 (陌生人 → 朋友 → 曖昧 → 戀人)
- 特殊事件觸發

### 6. API 設計

#### 6.1 對話相關
```
POST   /api/v1/chat/session              - 創建新會話
GET    /api/v1/chat/session/{id}         - 獲取會話資訊
GET    /api/v1/chat/sessions             - 獲取用戶會話列表 (支援分頁)
POST   /api/v1/chat/message              - 發送訊息
POST   /api/v1/chat/regenerate           - 重新生成回應
PUT    /api/v1/chat/session/{id}/mode    - 切換對話模式
GET    /api/v1/chat/session/{id}/history - 獲取會話對話歷史 (支援分頁)
POST   /api/v1/chat/session/{id}/tag     - 為會話添加標籤
GET    /api/v1/chat/session/{id}/export  - 匯出對話記錄
DELETE /api/v1/chat/session/{id}        - 結束對話
```

#### 6.2 角色相關
```
GET    /api/v1/character/list        - 獲取角色列表 (支援分頁與篩選)
GET    /api/v1/character/{id}        - 獲取角色詳細資訊
GET    /api/v1/character/{id}/stats  - 獲取角色統計數據
GET    /api/v1/user/character        - 獲取當前選擇角色
PUT    /api/v1/user/character        - 選擇當前角色
```

#### 6.3 小說模式
```
POST   /api/v1/novel/start             - 開始小說模式
POST   /api/v1/novel/choice            - 選擇劇情分支
POST   /api/v1/novel/progress/save     - 保存進度
GET    /api/v1/novel/progress/{id}     - 載入進度
GET    /api/v1/novel/progress/list     - 獲取存檔列表
```

#### 6.4 情感系統
```
GET    /api/v1/emotion/status        - 獲取情感狀態
GET    /api/v1/emotion/affection     - 獲取好感度
POST   /api/v1/emotion/event         - 觸發特殊事件
```

#### 6.5 TTS 相關
```
POST   /api/v1/tts/generate          - 生成語音
POST   /api/v1/tts/batch             - 批量生成語音
GET    /api/v1/tts/voices            - 獲取語音列表
POST   /api/v1/tts/preview           - 預覽語音
GET    /api/v1/tts/history           - 語音生成歷史
```

#### 6.6 記憶系統
```
GET    /api/v1/memory/user/{id}      - 獲取用戶記憶
POST   /api/v1/memory/save           - 手動保存記憶
DELETE /api/v1/memory/forget         - 選擇性遺忘
GET    /api/v1/memory/timeline       - 記憶時間線
POST   /api/v1/memory/search         - 搜尋記憶
```

#### 6.7 用戶系統
```
POST   /api/v1/user/register         - 用戶註冊
POST   /api/v1/user/login            - 用戶登入
POST   /api/v1/user/logout           - 用戶登出
POST   /api/v1/user/refresh          - 刷新 JWT Token
GET    /api/v1/user/profile          - 獲取個人資料
PUT    /api/v1/user/profile          - 更新個人資料
PUT    /api/v1/user/preferences      - 更新偏好設定
```

#### 6.8 系統管理
```
GET    /api/v1/health                - 健康檢查
GET    /api/v1/version               - API 版本
GET    /api/v1/status                - 系統狀態
```

#### 6.9 標籤系統
```
GET    /api/v1/tags                  - 獲取所有可用標籤
GET    /api/v1/tags/popular          - 獲取熱門標籤
```

#### API 設計說明
- **端點總數**: 68+ 個完整 API 端點
- **設計原則**: 符合 RESTful 規範，支援分頁、篩選、批量操作
- **認證方式**: JWT Bearer Token 統一認證
- **文檔格式**: 提供 OpenAPI 3.0 規格文件
- **詳細文檔**: 請參閱 [API.md](./API.md) 和 [openapi.yaml](./openapi.yaml)

### 7. 資料模型

#### 7.1 對話請求
```json
{
  "session_id": "string",
  "user_id": "string",
  "character_id": "string",
  "mode": "normal|novel|nsfw",
  "message": "string",
  "tags": ["string"],
  "context": {
    "affection": 0-100,
    "relationship": "stranger|friend|ambiguous|lover",
    "scene": "string"
  }
}
```

#### 7.2 對話回應
```json
{
  "session_id": "string",
  "character_id": "string",
  "response": "string",
  "emotion": "happy|sad|angry|shy|excited",
  "affection_change": -10 to +10,
  "engine_used": "openai|grok",
  "novel_choices": [
    {
      "id": "string",
      "text": "string",
      "consequence": "string"
    }
  ],
  "special_event": {
    "triggered": boolean,
    "type": "string",
    "description": "string"
  }
}
```

#### 7.3 角色資料
```json
{
  "id": "string",
  "name": "string",
  "type": "gentle|dominant|ascetic|sunny|cunning",
  "appearance": {
    "height": "string",
    "hair_color": "string",
    "eye_color": "string",
    "description": "string"
  },
  "personality": {
    "traits": ["string"],
    "likes": ["string"],
    "dislikes": ["string"]
  },
  "background": "string",
  "voice_id": "string",
  "initial_affection": 0-100
}
```

### 8. 系統配置

#### 8.1 內容策略
- **目標用戶**：成年 NSFW 內容接受者
- **內容限制**：無（完全開放 NSFW）
- **審核機制**：僅過濾違法內容

#### 8.2 資料管理
- **對話記錄**：所有用戶對話內容完整保存
- **存儲策略**：加密存儲所有內容（包含 NSFW）
- **清除選項**：用戶可主動選擇清除歷史
- **記憶保存**：永久保存所有互動記憶（除非用戶刪除）

#### 8.3 基本限制
- API 請求頻率限制（防止濫用）
- 並發會話數限制

### 9. 技術架構

#### 9.1 後端技術
- **語言**: Golang
- **框架**: Gin (HTTP), GORM (ORM)
- **資料庫**: PostgreSQL (用戶資料), Redis (會話快取), Qdrant (向量搜尋)
- **訊息隊列**: RabbitMQ (異步處理)

#### 9.2 AI 服務整合
- **OpenAI API**: GPT-4o (一般對話), TTS (語音)
- **Grok API**: NSFW 內容生成
- **內容審核**: 自建或第三方服務

#### 9.3 部署架構
- **容器化**: Docker
- **部署**: Docker Compose（開發環境）
- **負載均衡**: Nginx
- **監控**: Prometheus + Grafana

### 10. 開發計劃

#### Phase 1: 基礎建設 (Week 1)
- [ ] 專案架構設計
- [ ] 資料庫設計 (PostgreSQL + Redis + Qdrant)
- [ ] Gin API 框架搭建
- [ ] 基礎配置管理 (Viper)
- [ ] 健康檢查與系統 API

#### Phase 2: 用戶與會話 (Week 2)
- [ ] 用戶註冊/登入系統
- [ ] JWT 認證機制
- [ ] 會話創建與管理
- [ ] 基礎 CRUD API

#### Phase 3: 核心對話 (Week 3-4)
- [ ] OpenAI GPT-4o 整合
- [ ] 基本對話功能
- [ ] 兩個預設角色實作
- [ ] 對話上下文管理
- [ ] 對話歷史記錄

#### Phase 4: 記憶系統 (Week 5-6)
- [ ] Redis 短期記憶實作
- [ ] PostgreSQL 長期記憶
- [ ] Qdrant 向量搜尋整合
- [ ] 記憶檢索機制
- [ ] 上下文組裝邏輯
- [ ] 記憶管理 API

#### Phase 5: 模式系統 (Week 7-8)
- [ ] 普通對話模式
- [ ] 小說模式框架
- [ ] NSFW 模式切換
- [ ] 標籤系統實作
- [ ] 模式路由邏輯

#### Phase 6: NSFW 功能 (Week 9-10)
- [ ] Grok API 整合
- [ ] NSFW 內容生成
- [ ] 標籤觸發機制
- [ ] 內容路由系統
- [ ] 內容分類器

#### Phase 7: 情感系統 (Week 11-12)
- [ ] 好感度計算邏輯
- [ ] 關係狀態管理
- [ ] 特殊事件系統
- [ ] 情感記憶整合
- [ ] 情感變化追蹤

#### Phase 8: 語音功能 (Week 13)
- [ ] OpenAI TTS 整合
- [ ] 兩個角色語音配置
- [ ] 語音生成 API
- [ ] 音頻檔案管理

#### Phase 9: 測試優化 (Week 14)
- [ ] 單元測試
- [ ] 整合測試
- [ ] 記憶系統優化
- [ ] API 效能調整
- [ ] Docker 容器化


### 11. 待確認事項

1. **記憶系統細節**
   - Qdrant 集群配置？
   - 記憶容量上限？
   - 記憶衝突處理策略？

2. **技術細節**
   - API 認證方式（JWT/API Key）？
   - 日誌保留策略？
   - 連接池大小設定？

## 重要提醒
本規格書為初版，所有功能細節需經過討論確認後才進行實作。請仔細審查每個模組，確保符合產品願景和用戶需求。