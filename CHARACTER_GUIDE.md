# 角色系統設計指南（Character Guide）

本指南說明角色（Character）在對話系統中的定位、資料與配置結構、選擇策略（風格/場景/狀態/NSFW）、存取權限與快照/快取原則，作為產品與工程溝通的共同基準。內容聚焦方法與原則，不包含程式碼細節。

## 目標與原則

- 一致人格：在不同情境與語系下保持連貫且可預期的人設表現。
- 情境適配：基於好感度、NSFW 等級、時間段、場景等上下文調整語氣與內容。
- 可維運：角色資訊集中管理、可版本化、可審計，具備快取與快照以支援高併發讀取。
- 最小驚訝：公開端點只暴露必要資料；管理端點明確權限與變更流程。

## 核心概念

- 角色（Character）：人物的核心資料與行為總和（名稱、描述、類型、語系、啟用狀態）。
- 檔案（Profile）：背景敘事、外觀、性格特質（Traits / Values / Scores）。
- 行為（Behavior）：
  - 對話風格（Speech Styles）：不同語氣/長度/詞彙偏好，含權重、好感範圍與 NSFW 范圍。
  - NSFW 設定（NSFW Config/Levels）：等級對應的引擎、指引與詞彙邊界。
  - 情緒設定（Emotional Config）：預設心情、支援情緒、觸發因子與變化範圍。
  - 互動規則（Interaction Rules）：跨場景應遵循的原則。
- 內容（Content）：
  - 場景（Scenes）：可被挑選的場景片段，受時間段/好感度/NSFW 影響。
  - 狀態（States）：人物當前狀態（如工作/關懷/深情）可加權挑選。
- 本地化（Localization）：多語系名稱與敘述、職業、年齡等在地化資訊，支援 fallback。
- 快照（Snapshot）：將完整角色聚合序列化為可快速讀取的只讀視圖，供聊天熱路徑使用。

## 權限與端點（建議）

- 公開（無需登入）：
  - 角色列表、詳情、統計、NSFW 指引（依等級）。
- 使用者（登入）：
  - 讀取角色檔案與配置摘要、依上下文取得最佳對話風格/場景/狀態。
- 管理員：
  - 角色 CRUD、風格與場景/狀態管理、刷新快照/快取。

## 概念資料結構（邏輯層）

- Core：Character（id, name, type, locale, is_active, avatar, popularity, tags, timestamps）
- Profile：description, background, appearance, personality（traits/values/strengths/weaknesses/scores）
- Behavior：
  - Speech Styles：name, type, tone, length range, positive/negative keywords, templates, weight,
    affection_range, nsfw_range, active
  - NSFW：max_level, require_adult_age, restrictions；Levels：level, engine, title, desc, guidelines,
    positive/negative keywords, temperature, active
  - Emotional：default_mood, supported_moods, triggers, variability, affection range
  - Interaction Rules：rule list
- Content：Scenes（type, time_of_day, desc, affection & nsfw range, weight, active）、States（key, desc, affection range, weight, active）
- Localization：per-locale name/description/background/profession/age
- Snapshot：完整聚合 JSON，含版本與刷新時間

## 選擇策略（運行時）

- 對話風格（Speech Style）：
  - 條件過濾：is_active、nsfw 在範圍內、affection 在範圍內。
  - 競爭策略：採權重最高（或權重抽樣）且符合當前語系與上下文者。
  - 缺省策略：篩不到時回退至預設風格。

- 場景（Scene）：
  - 條件過濾：is_active、time_of_day、scene_type、affection、nsfw。
  - 排序策略：以 weight 為主、可混入新鮮度/最近使用衰減。

- 狀態（State）：
- 條件過濾：is_active、affection 範圍。
  - 排序策略：以 weight 為主。

- 本地化（Localization）：
  - 對目標 locale 取值；若缺失則回退至角色預設語系。

- NSFW 指引：
  - 依等級返回標題/描述/指引與關鍵詞；成人等級需同意與年齡驗證。

## 快照與快取

- 快照（Read Model）：
  - 用途：聊天熱路徑與搜尋；避免多表組裝帶來的延遲。
  - 刷新：角色 CRUD 或子配置變更後刷新；支援全量/單角色刷新端點。
  - 版本：快照帶版本號與時間戳，支援條件查詢（If-None-Match/ETag）。

- 快取：
  - 進程內或集中式快取（按部署架構選擇）存放快照或部分子配置；TTL 短期 + 事件失效。
  - Key 建議：`character:snapshot:{id}:v{version}`。

## 版本管理與變更治理

- 角色資料的變更（描述、風格、場景等）需記錄審計（操作者、時間、差異）。
- 大幅變更或下架：
  - 對現有會話僅影響後續輪次；保留舊回應的歸檔一致性。
  - 善用 `is_active` 作為下架開關，避免硬刪除破壞引用。

## 多語系與在地化

- 最小集合：`zh-Hant`（繁）、`zh-Hans`（簡）、`ja`、`en`。
- Fallback 規則：未提供某語系時回退至角色預設語系；避免混雜語言敘述。
- 在地化內容僅覆蓋必要欄位（name/description/background/profession/age），其餘使用通用設定。

## 一致性與品控

- 一致性檢查：
  - 說話風格：句長、命令/溫柔等語氣、避免違反角色限制。
  - 性格特質：是否體現核心價值與典型行為。
  - 詞彙選擇：正負面詞彙白名單/黑名單。
- 不一致處理：
  - 記錄違規類型與嚴重度；必要時二次生成或微調提示詞。

## NSFW 方針（等級建議）

- 1 安全（Safe）：日常對話與輕微曖昧，避免任何露骨。
- 2 浪漫（Romantic）：情感細節增強，可有含蓄的親密描述。
- 3 親密（Intimate）：可有直接身體接觸描寫，但需優雅與尊重。
- 4 成人（Adult）：成人向內容但保留美感與同意，嚴禁暴力/羞辱等。
- 5 明確（Explicit）：更直接的成人描述，但不得出現任何違規或不尊重內容。

## 觀測與 SLO

- 指標：角色快照命中率、讀取延遲、風格選擇分佈、場景使用分佈、錯誤率。
- 錯誤碼：角色不存在、驗證失敗、配置缺失、快照未刷新、權限不足等。
- SLO 建議：角色詳情 P95 < 50ms（快照命中）、P99 < 100ms；刷新在數秒內完成。

## 驗收清單

- [x] 角色列表/詳情可公開檢索，返回在地化後的精簡資料。
- [x] 依 NSFW 等級、好感度能選到合理的對話風格（缺省有回退）。
- [x] 場景與狀態依條件可被挑選且具可解釋的權重邏輯。
- [x] 快照/快取可手動刷新並帶版本資訊。
- [x] 管理端修改有審計紀錄；停用不破壞既有會話。
- [x] 一致性檢查覆蓋語氣/性格/詞彙三層，並提供建議。

## 後續擴展

- 角色 A/B：同一人設多個語氣配方測試，收斂最優權重。
- 自適應風格：結合互動反饋自動調整權重與選擇策略。
- 內容生長：由使用者互動驅動場景/狀態庫的增長與季節性更新。

