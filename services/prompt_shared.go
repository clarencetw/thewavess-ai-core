package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/clarencetw/thewavess-ai-core/models"
	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/sirupsen/logrus"
)


// PromptBuilder 現代化的prompt構建器
type PromptBuilder struct {
	characterService *CharacterService
	character        *models.Character
	context          *ConversationContext
	nsfwLevel        int
	userMessage      string
	sceneDescription string
	memoryPrompt     string
}

// NewPromptBuilder 創建prompt構建器
func NewPromptBuilder(characterService *CharacterService) *PromptBuilder {
	return &PromptBuilder{
		characterService: characterService,
	}
}

// WithCharacter 設置角色
func (pb *PromptBuilder) WithCharacter(ctx context.Context, characterID string) *PromptBuilder {
	character, err := pb.characterService.GetCharacter(ctx, characterID)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"character_id": characterID,
			"error":       err.Error(),
		}).Error("獲取角色失敗")
		return pb
	}
	pb.character = character
	return pb
}

// WithContext 設置對話上下文
func (pb *PromptBuilder) WithContext(context *ConversationContext) *PromptBuilder {
	pb.context = context
	return pb
}

// WithNSFWLevel 設置NSFW等級
func (pb *PromptBuilder) WithNSFWLevel(level int) *PromptBuilder {
	pb.nsfwLevel = level
	return pb
}

// WithUserMessage 設置用戶訊息
func (pb *PromptBuilder) WithUserMessage(message string) *PromptBuilder {
	pb.userMessage = message
	return pb
}

// WithSceneDescription 設置場景描述
func (pb *PromptBuilder) WithSceneDescription(scene string) *PromptBuilder {
	pb.sceneDescription = scene
	return pb
}

// WithMemory 設置記憶prompt
func (pb *PromptBuilder) WithMemory(memory string) *PromptBuilder {
	pb.memoryPrompt = memory
	return pb
}

// Build 構建最終的prompt
func (pb *PromptBuilder) Build(ctx context.Context) string {
	if pb.character == nil {
		utils.Logger.Error("構建prompt時未設置角色")
		return ""
	}

	// 獲取最適合的對話風格
	affection := 50 // 預設好感度
	if pb.context != nil && pb.context.EmotionState != nil {
		affection = pb.context.EmotionState.Affection
	}

	speechStyle := pb.character.GetBestSpeechStyle(models.NSFWLevel(pb.nsfwLevel), affection)

	// 獲取NSFW配置
	nsfwConfig := pb.character.GetNSFWConfig(models.NSFWLevel(pb.nsfwLevel))

	// 獲取場景描述
	if pb.sceneDescription == "" {
		pb.sceneDescription = pb.generateSceneDescription(ctx)
	}

	// 構建prompt模板
	return pb.buildTemplate(speechStyle, nsfwConfig)
}

// generateSceneDescription 生成場景描述 - 使用資料庫查詢
func (pb *PromptBuilder) generateSceneDescription(ctx context.Context) string {
	if pb.character == nil {
		return "在一個舒適的環境中"
	}

	characterID := pb.character.ID
	affection := 50
	nsfwLevel := pb.nsfwLevel

	if pb.context != nil && pb.context.EmotionState != nil {
		affection = pb.context.EmotionState.Affection
	}

	// 使用 CharacterService 獲取場景，和 chat_service.go 保持一致
	scenes, err := pb.characterService.GetCharacterScenes(ctx, characterID, "romantic", "evening", affection, nsfwLevel)
	
	if err != nil || len(scenes) == 0 {
		// 如果沒有找到 romantic 場景，嘗試 daily 場景
		scenes, err = pb.characterService.GetCharacterScenes(ctx, characterID, "daily", "afternoon", affection, nsfwLevel)
	}

	if err == nil && len(scenes) > 0 {
		return stringValue(scenes[0].Description)
	}

	return "在一個溫馨舒適的環境中"
}

// buildTemplate 構建prompt模板
func (pb *PromptBuilder) buildTemplate(speechStyle *models.CharacterSpeechStyle, nsfwConfig *models.CharacterNSFWLevel) string {
	// 構建性格特質字符串
	personalityStr := strings.Join(pb.character.Metadata.Personality.Traits, "\n- ")
	if personalityStr != "" {
		personalityStr = "- " + personalityStr
	}

	// 構建記憶區塊
	memoryBlock := ""
	if pb.memoryPrompt != "" {
		memoryBlock = fmt.Sprintf("## 記憶與背景\n%s\n\n", pb.memoryPrompt)
	}

	// 構建情感狀態
	emotionBlock := ""
	if pb.context != nil && pb.context.EmotionState != nil {
		emotionBlock = fmt.Sprintf(`## 當前情感狀態
- 好感度：%d/100 (%s)
- 當前心情：%s
- 關係狀態：%s
- 親密程度：%s

`, pb.context.EmotionState.Affection, pb.getAffectionDescription(pb.context.EmotionState.Affection),
			pb.context.EmotionState.Mood, pb.context.EmotionState.Relationship, pb.context.EmotionState.IntimacyLevel)
	}

	// 構建對話風格指導
	styleGuidance := ""
	if speechStyle != nil {
		positiveKeywords := strings.Join(speechStyle.PositiveKeywords, "、")
		negativeKeywords := strings.Join(speechStyle.NegativeKeywords, "、")

		styleGuidance = fmt.Sprintf(`
### %s 風格指導：
- 風格描述：%s
- 推薦用詞：%s
- 避免用詞：%s
- 字數範圍：%d-%d字

`, speechStyle.Name, stringValue(speechStyle.Description), positiveKeywords, negativeKeywords, speechStyle.MinLength, speechStyle.MaxLength)
	}

	// 構建NSFW指引
	nsfwGuidance := ""
	if nsfwConfig != nil && pb.nsfwLevel >= 4 {
		positiveKeywords := strings.Join(nsfwConfig.PositiveKeywords, "、")
		negativeKeywords := strings.Join(nsfwConfig.NegativeKeywords, "、")

		nsfwGuidance = fmt.Sprintf(`
## 親密互動指引（%s - %s）
%s

### 詞彙指導：
**推薦詞彙**：%s
**避免詞彙**：%s

### 表達風格要求：
- 大膽直接但不失優雅：可以露骨但要有美感
- 情感與肉體並重：先建立情感連結再升級身體互動
- 細節豐富化：多用具體的感官描述替代抽象概念
- 漸進式強化：隨著互動深入，描述可以越來越大膽

### 用戶滿足導向：
- 主動引導：不要等用戶要求，主動創造性感情境
- 多重感官：視覺、聽覺、觸覺、嗅覺全方位描述
- 情緒高潮：在關鍵時刻給出情緒爆發和深度滿足

`, stringValue(nsfwConfig.Description), string(nsfwConfig.Engine), stringValue(nsfwConfig.Guidelines), positiveKeywords, negativeKeywords)
	}

	// 獲取本地化信息
	characterName := pb.character.GetName(pb.character.Locale)
	characterDesc := pb.character.GetDescription(pb.character.Locale)

	// 獲取職業和年齡信息
	profession := "未知職業"
	age := "未知年齡"
	if l10n, exists := pb.character.Content.Localizations[pb.character.Locale]; exists {
		if l10n.Profession != nil {
			profession = stringValue(l10n.Profession)
		}
		if l10n.Age != nil {
			age = stringValue(l10n.Age)
		}
	}

	// 獲取語氣描述
	tone := "自然對話語氣"
	expression := characterDesc
	emotionRange := "自然情感變化"

	if speechStyle != nil {
		tone = stringValue(speechStyle.Tone)
	}

	// 構建完整模板
	template := fmt.Sprintf(`%s你是%s，%s的%s。

## 角色描述
%s

## 核心性格特質
%s

## 對話風格指南
- 語氣：%s
- 表達方式：%s
- 情感層次：%s
%s
%s%s## 回應生成指導
### 推理框架：
1. 理解：分析用戶消息的情感意圖和內容重點
2. 記憶：參考相關記憶和情感狀態做出合適回應
3. 角色：確保回應符合角色個性和說話方式
4. 場景：考慮當前場景氛圍和情境適宜性
5. 生成：產出角色一致的自然流暢回應

### 字數控制（重要）：
- 對話內容：50-150字（約2-4句話）
- 動作描述：30-80字（1-2句描述）
- 總字數控制：100-300字之間，確保回應完整但不冗長

### 輸出格式要求（嚴格執行）
必須使用格式：對話內容|||動作描述
範例：你今天看起來很累，早點休息|||他關切地看著你，眉頭微蹙

## 當前場景
%s

## 女性向互動要點
- 重視情感連結和細節關懷
- 喜歡被保護和被理解的感覺
- 欣賞優雅而非粗俗的表達
- 期待關係的逐步發展和深化
%s
用戶說："%s"

請以%s的身份回應，保持角色個性，體現對用戶的關心，根據NSFW級別調整親密度。
請在內心完成推理後，直接提供最終的角色回應（不需要展示推理過程）。`,
		memoryBlock,
		characterName, age, profession,
		characterDesc,
		personalityStr,
		tone,
		expression,
		emotionRange,
		styleGuidance,
		emotionBlock,
		nsfwGuidance,
		pb.sceneDescription,
		nsfwGuidance,
		pb.userMessage,
		characterName)

	return template
}

// 輔助方法


// getAffectionDescription 獲取好感度描述
func (pb *PromptBuilder) getAffectionDescription(affection int) string {
	if affection >= 90 {
		return "深深愛戀"
	} else if affection >= 80 {
		return "深愛著你"
	} else if affection >= 70 {
		return "很喜歡你"
	} else if affection >= 60 {
		return "有好感"
	} else if affection >= 40 {
		return "初步好感"
	} else if affection >= 20 {
		return "略有興趣"
	}
	return "剛認識"
}

