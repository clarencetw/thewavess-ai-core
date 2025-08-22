package models

import (
	"encoding/json"
	"time"
	
	"github.com/google/uuid"
	"github.com/clarencetw/thewavess-ai-core/models/db"
)

// CharacterMapper 角色模型轉換器
type CharacterMapper struct{}

// NewCharacterMapper 創建新的角色映射器
func NewCharacterMapper() *CharacterMapper {
	return &CharacterMapper{}
}

// FromDB 從資料庫模型轉換為領域模型
func (m *CharacterMapper) FromDB(
	charDB *db.CharacterDB,
	profileDB *db.CharacterProfileDB,
	localizationsDB []*db.CharacterLocalizationDB,
	speechStylesDB []*db.CharacterSpeechStyleDB,
	scenesDB []*db.CharacterSceneDB,
	statesDB []*db.CharacterStateDB,
	emotionalConfigDB *db.CharacterEmotionalConfigDB,
	nsfwConfigDB *db.CharacterNSFWConfigDB,
	nsfwLevelsDB []*db.CharacterNSFWLevelDB,
	interactionRulesDB []*db.CharacterInteractionRuleDB,
) (*Character, error) {
	character := &Character{
		ID:        charDB.ID,
		Name:      charDB.Name,
		Type:      CharacterType(charDB.Type),
		Locale:    Locale(charDB.Locale),
		IsActive:  charDB.IsActive,
		CreatedAt: charDB.CreatedAt,
		UpdatedAt: charDB.UpdatedAt,
	}

	// Metadata
	character.Metadata = CharacterMetadata{
		Tags:       charDB.Tags,
		Popularity: charDB.Popularity,
	}
	
	if charDB.AvatarURL != nil {
		character.Metadata.AvatarURL = charDB.AvatarURL
	}

	if profileDB != nil {
		character.Metadata.Description = &profileDB.Description
		if profileDB.Background != nil {
			character.Metadata.Background = profileDB.Background
		}
		
		// 轉換 Appearance
		if profileDB.Appearance != nil {
			if appearance, err := m.mapAppearance(profileDB.Appearance); err == nil {
				character.Metadata.Appearance = appearance
			}
		}

		// 轉換 Personality
		if profileDB.Personality != nil {
			if personality, err := m.mapPersonality(profileDB.Personality); err == nil {
				character.Metadata.Personality = personality
			}
		}
	}

	// Localizations
	character.Content.Localizations = make(map[Locale]CharacterL10N)
	for _, locDB := range localizationsDB {
		loc := CharacterL10N{}
		if locDB.Name != nil {
			loc.Name = locDB.Name
		}
		if locDB.Description != nil {
			loc.Description = locDB.Description
		}
		if locDB.Background != nil {
			loc.Background = locDB.Background
		}
		if locDB.Profession != nil {
			loc.Profession = locDB.Profession
		}
		if locDB.Age != nil {
			loc.Age = locDB.Age
		}
		character.Content.Localizations[Locale(locDB.Locale)] = loc
	}

	// Speech Styles
	character.Behavior.SpeechStyles = make([]CharacterSpeechStyle, len(speechStylesDB))
	for i, styleDB := range speechStylesDB {
		style := CharacterSpeechStyle{
			Name:             styleDB.Name,
			StyleType:        StyleType(styleDB.StyleType),
			MinLength:        styleDB.MinLength,
			MaxLength:        styleDB.MaxLength,
			PositiveKeywords: styleDB.PositiveKeywords,
			NegativeKeywords: styleDB.NegativeKeywords,
			Templates:        styleDB.Templates,
			Weight:           styleDB.Weight,
			IsActive:         styleDB.IsActive,
			AffectionRange:   [2]int{styleDB.AffectionMin, styleDB.AffectionMax},
			NSFWRange:        [2]int{styleDB.NSFWMin, styleDB.NSFWMax},
		}
		if styleDB.Tone != nil {
			style.Tone = styleDB.Tone
		}
		if styleDB.Description != nil {
			style.Description = styleDB.Description
		}
		character.Behavior.SpeechStyles[i] = style
	}

	// Scenes
	character.Content.Scenes = make([]CharacterScene, len(scenesDB))
	for i, sceneDB := range scenesDB {
		scene := CharacterScene{
			ID:       sceneDB.ID,
			IsActive: sceneDB.IsActive,
		}
		if sceneDB.SceneType != nil {
			scene.SceneType = sceneDB.SceneType
		}
		if sceneDB.TimeOfDay != nil {
			scene.TimeOfDay = sceneDB.TimeOfDay
		}
		if sceneDB.Description != nil {
			scene.Description = sceneDB.Description
		}
		if sceneDB.AffectionMin != nil {
			scene.AffectionMin = *sceneDB.AffectionMin
		}
		if sceneDB.AffectionMax != nil {
			scene.AffectionMax = *sceneDB.AffectionMax
		}
		if sceneDB.NSFWLevelMin != nil {
			scene.NSFWLevelMin = NSFWLevel(*sceneDB.NSFWLevelMin)
		}
		if sceneDB.NSFWLevelMax != nil {
			scene.NSFWLevelMax = NSFWLevel(*sceneDB.NSFWLevelMax)
		}
		if sceneDB.Weight != nil {
			scene.Weight = *sceneDB.Weight
		}
		character.Content.Scenes[i] = scene
	}

	// States
	character.Content.States = make([]CharacterState, len(statesDB))
	for i, stateDB := range statesDB {
		state := CharacterState{
			ID:       stateDB.ID,
			IsActive: stateDB.IsActive,
		}
		if stateDB.StateKey != nil {
			state.StateKey = stateDB.StateKey
		}
		if stateDB.Description != nil {
			state.Description = stateDB.Description
		}
		if stateDB.AffectionMin != nil {
			state.AffectionMin = *stateDB.AffectionMin
		}
		if stateDB.AffectionMax != nil {
			state.AffectionMax = *stateDB.AffectionMax
		}
		if stateDB.Weight != nil {
			state.Weight = *stateDB.Weight
		}
		character.Content.States[i] = state
	}

	// Emotional Config
	if emotionalConfigDB != nil {
		emotional := CharacterEmotionalConfig{}
		emotional.DefaultMood = emotionalConfigDB.DefaultMood
		emotional.EmotionRange = emotionalConfigDB.EmotionRange
		if emotionalConfigDB.AffectionStart != nil {
			emotional.AffectionStart = *emotionalConfigDB.AffectionStart
		}
		if emotionalConfigDB.MaxAffection != nil {
			emotional.MaxAffection = *emotionalConfigDB.MaxAffection
		}
		if emotionalConfigDB.MoodVariability != nil {
			emotional.MoodVariability = *emotionalConfigDB.MoodVariability
		}
		emotional.SupportedMoods = emotionalConfigDB.SupportedMoods
		emotional.EmotionalTriggers = emotionalConfigDB.EmotionalTriggers
		character.Behavior.EmotionalConfig = emotional
	}

	// NSFW Config  
	if nsfwConfigDB != nil {
		nsfw := CharacterNSFWConfig{}
		if nsfwConfigDB.MaxLevel != nil {
			nsfw.MaxLevel = NSFWLevel(*nsfwConfigDB.MaxLevel)
		}
		if nsfwConfigDB.RequireAdultAge != nil {
			nsfw.RequireAdultAge = *nsfwConfigDB.RequireAdultAge
		}
		nsfw.Restrictions = nsfwConfigDB.Restrictions
		character.Behavior.NSFWConfig = nsfw
	}

	// NSFW Levels - 放在 NSFWConfig 內部
	nsfwLevels := make([]CharacterNSFWLevel, len(nsfwLevelsDB))
	for i, levelDB := range nsfwLevelsDB {
		level := CharacterNSFWLevel{
			Level:            NSFWLevel(levelDB.Level),
			Engine:           EngineType(levelDB.Engine),
			PositiveKeywords: levelDB.PositiveKeywords,
			NegativeKeywords: levelDB.NegativeKeywords,
			IsActive:         levelDB.IsActive,
		}
		level.Title = levelDB.Title
		level.Description = levelDB.Description
		level.Guidelines = levelDB.Guidelines
		if levelDB.Temperature != nil {
			temp := float64(*levelDB.Temperature)
			level.Temperature = &temp
		}
		nsfwLevels[i] = level
	}
	// 將 levels 添加到現有的 NSFWConfig 中
	if len(nsfwLevels) > 0 {
		character.Behavior.NSFWConfig.LevelConfigs = nsfwLevels
	}

	// Interaction Rules
	character.Behavior.InteractionRules = make([]string, len(interactionRulesDB))
	for i, ruleDB := range interactionRulesDB {
		character.Behavior.InteractionRules[i] = ruleDB.Rule
	}

	return character, nil
}

// ToDB 從領域模型轉換為資料庫模型群組
func (m *CharacterMapper) ToDB(character *Character) (*CharacterDBGroup, error) {
	group := &CharacterDBGroup{}

	// 基礎角色
	group.Character = &db.CharacterDB{
		ID:         character.ID,
		Name:       character.Name,
		Type:       string(character.Type),
		Locale:     string(character.Locale),
		IsActive:   character.IsActive,
		Popularity: character.Metadata.Popularity,
		Tags:       character.Metadata.Tags,
		CreatedAt:  character.CreatedAt,
		UpdatedAt:  character.UpdatedAt,
	}

	group.Character.AvatarURL = character.Metadata.AvatarURL

	// 角色檔案
	group.Profile = &db.CharacterProfileDB{
		CharacterID: character.ID,
		Description: *character.Metadata.Description,
		UpdatedAt:   time.Now(),
	}

	group.Profile.Background = character.Metadata.Background

	// 轉換 Appearance 和 Personality
	if len(character.Metadata.Appearance.Height) > 0 || len(character.Metadata.Appearance.Build) > 0 {
		appearanceData, _ := json.Marshal(character.Metadata.Appearance)
		var appearance map[string]interface{}
		json.Unmarshal(appearanceData, &appearance)
		group.Profile.Appearance = appearance
	}

	if len(character.Metadata.Personality.Traits) > 0 || len(character.Metadata.Personality.CoreValues) > 0 {
		personalityData, _ := json.Marshal(character.Metadata.Personality)
		var personality map[string]interface{}
		json.Unmarshal(personalityData, &personality)
		group.Profile.Personality = personality
	}

	// 本地化
	for locale, loc := range character.Content.Localizations {
		locDB := &db.CharacterLocalizationDB{
			CharacterID: character.ID,
			Locale:      string(locale),
		}
		locDB.Name = loc.Name
		locDB.Description = loc.Description
		locDB.Background = loc.Background
		locDB.Profession = loc.Profession
		locDB.Age = loc.Age
		group.Localizations = append(group.Localizations, locDB)
	}

	// Speech Styles
	for _, style := range character.Behavior.SpeechStyles {
		styleDB := &db.CharacterSpeechStyleDB{
			ID:               "style_" + uuid.New().String(),
			CharacterID:      character.ID,
			Name:             style.Name,
			StyleType:        string(style.StyleType),
			MinLength:        style.MinLength,
			MaxLength:        style.MaxLength,
			PositiveKeywords: style.PositiveKeywords,
			NegativeKeywords: style.NegativeKeywords,
			Templates:        style.Templates,
			Weight:           style.Weight,
			IsActive:         style.IsActive,
			AffectionMin:     style.AffectionRange[0],
			AffectionMax:     style.AffectionRange[1],
			NSFWMin:          style.NSFWRange[0],
			NSFWMax:          style.NSFWRange[1],
		}
		styleDB.Tone = style.Tone
		styleDB.Description = style.Description
		group.SpeechStyles = append(group.SpeechStyles, styleDB)
	}

	// Scenes
	for _, scene := range character.Content.Scenes {
		sceneDB := &db.CharacterSceneDB{
			ID:          "scene_" + uuid.New().String(),
			CharacterID: character.ID,
			IsActive:    scene.IsActive,
		}
		sceneDB.SceneType = scene.SceneType
		sceneDB.TimeOfDay = scene.TimeOfDay
		sceneDB.Description = scene.Description
		if scene.AffectionMin > 0 {
			sceneDB.AffectionMin = &scene.AffectionMin
		}
		if scene.AffectionMax > 0 {
			sceneDB.AffectionMax = &scene.AffectionMax
		}
		if scene.NSFWLevelMin > 0 {
			nsfwMin := int(scene.NSFWLevelMin)
			sceneDB.NSFWLevelMin = &nsfwMin
		}
		if scene.NSFWLevelMax > 0 {
			nsfwMax := int(scene.NSFWLevelMax)
			sceneDB.NSFWLevelMax = &nsfwMax
		}
		if scene.Weight > 0 {
			sceneDB.Weight = &scene.Weight
		}
		group.Scenes = append(group.Scenes, sceneDB)
	}

	// States
	for _, state := range character.Content.States {
		stateDB := &db.CharacterStateDB{
			ID:          "state_" + uuid.New().String(),
			CharacterID: character.ID,
			IsActive:    state.IsActive,
		}
		stateDB.StateKey = state.StateKey
		stateDB.Description = state.Description
		if state.AffectionMin > 0 {
			stateDB.AffectionMin = &state.AffectionMin
		}
		if state.AffectionMax > 0 {
			stateDB.AffectionMax = &state.AffectionMax
		}
		if state.Weight > 0 {
			stateDB.Weight = &state.Weight
		}
		group.States = append(group.States, stateDB)
	}

	// Emotional Config
	group.EmotionalConfig = &db.CharacterEmotionalConfigDB{
		CharacterID: character.ID,
	}
	group.EmotionalConfig.DefaultMood = character.Behavior.EmotionalConfig.DefaultMood
	group.EmotionalConfig.EmotionRange = character.Behavior.EmotionalConfig.EmotionRange
	if character.Behavior.EmotionalConfig.AffectionStart > 0 {
		group.EmotionalConfig.AffectionStart = &character.Behavior.EmotionalConfig.AffectionStart
	}
	if character.Behavior.EmotionalConfig.MaxAffection > 0 {
		group.EmotionalConfig.MaxAffection = &character.Behavior.EmotionalConfig.MaxAffection
	}
	if character.Behavior.EmotionalConfig.MoodVariability > 0 {
		group.EmotionalConfig.MoodVariability = &character.Behavior.EmotionalConfig.MoodVariability
	}
	group.EmotionalConfig.SupportedMoods = character.Behavior.EmotionalConfig.SupportedMoods
	group.EmotionalConfig.EmotionalTriggers = character.Behavior.EmotionalConfig.EmotionalTriggers

	// NSFW Config
	group.NSFWConfig = &db.CharacterNSFWConfigDB{
		CharacterID: character.ID,
	}
	if character.Behavior.NSFWConfig.MaxLevel > 0 {
		maxLevel := int(character.Behavior.NSFWConfig.MaxLevel)
		group.NSFWConfig.MaxLevel = &maxLevel
	}
	if character.Behavior.NSFWConfig.RequireAdultAge {
		group.NSFWConfig.RequireAdultAge = &character.Behavior.NSFWConfig.RequireAdultAge
	}
	group.NSFWConfig.Restrictions = character.Behavior.NSFWConfig.Restrictions

	// NSFW Levels
	for _, level := range character.Behavior.NSFWConfig.LevelConfigs {
		levelDB := &db.CharacterNSFWLevelDB{
			ID:               "level_" + uuid.New().String(),
			CharacterID:      character.ID,
			Level:            int(level.Level),
			Engine:           string(level.Engine),
			PositiveKeywords: level.PositiveKeywords,
			NegativeKeywords: level.NegativeKeywords,
			IsActive:         level.IsActive,
		}
		levelDB.Title = level.Title
		levelDB.Description = level.Description
		levelDB.Guidelines = level.Guidelines
		if level.Temperature != nil {
			temp := float32(*level.Temperature)
			levelDB.Temperature = &temp
		}
		group.NSFWLevels = append(group.NSFWLevels, levelDB)
	}

	// Interaction Rules
	for _, rule := range character.Behavior.InteractionRules {
		ruleDB := &db.CharacterInteractionRuleDB{
			ID:          "rule_" + uuid.New().String(),
			CharacterID: character.ID,
			Rule:        rule,
		}
		group.InteractionRules = append(group.InteractionRules, ruleDB)
	}

	return group, nil
}

// mapAppearance 轉換外觀資料
func (m *CharacterMapper) mapAppearance(data map[string]interface{}) (CharacterAppearance, error) {
	appearance := CharacterAppearance{}
	
	if height, ok := data["height"].(string); ok {
		appearance.Height = height
	}
	if build, ok := data["build"].(string); ok {
		appearance.Build = build
	}
	if eyeColor, ok := data["eye_color"].(string); ok {
		appearance.EyeColor = eyeColor
	}
	if hairColor, ok := data["hair_color"].(string); ok {
		appearance.HairColor = hairColor
	}
	if hairStyle, ok := data["hair_style"].(string); ok {
		appearance.HairStyle = hairStyle
	}
	if skinTone, ok := data["skin_tone"].(string); ok {
		appearance.SkinTone = skinTone
	}
	if style, ok := data["style"].(string); ok {
		appearance.Style = style
	}
	if distinctive, ok := data["distinctive"].(string); ok {
		appearance.Distinctive = distinctive
	}

	return appearance, nil
}

// mapPersonality 轉換性格資料
func (m *CharacterMapper) mapPersonality(data map[string]interface{}) (CharacterPersonality, error) {
	personality := CharacterPersonality{}

	if traits, ok := data["traits"].([]interface{}); ok {
		for _, trait := range traits {
			if traitStr, ok := trait.(string); ok {
				personality.Traits = append(personality.Traits, traitStr)
			}
		}
	}

	if coreValues, ok := data["core_values"].([]interface{}); ok {
		for _, value := range coreValues {
			if valueStr, ok := value.(string); ok {
				personality.CoreValues = append(personality.CoreValues, valueStr)
			}
		}
	}

	if strengths, ok := data["strengths"].([]interface{}); ok {
		for _, strength := range strengths {
			if strengthStr, ok := strength.(string); ok {
				personality.Strengths = append(personality.Strengths, strengthStr)
			}
		}
	}

	if weaknesses, ok := data["weaknesses"].([]interface{}); ok {
		for _, weakness := range weaknesses {
			if weaknessStr, ok := weakness.(string); ok {
				personality.Weaknesses = append(personality.Weaknesses, weaknessStr)
			}
		}
	}

	if scoreData, ok := data["personality_score"].(map[string]interface{}); ok {
		score := CharacterPersonalityScore{}
		if extroversion, ok := scoreData["extroversion"].(float64); ok {
			score.Extroversion = int(extroversion)
		}
		if agreeableness, ok := scoreData["agreeableness"].(float64); ok {
			score.Agreeableness = int(agreeableness)
		}
		if dominance, ok := scoreData["dominance"].(float64); ok {
			score.Dominance = int(dominance)
		}
		if emotional, ok := scoreData["emotional"].(float64); ok {
			score.Emotional = int(emotional)
		}
		if openness, ok := scoreData["openness"].(float64); ok {
			score.Openness = int(openness)
		}
		if reliability, ok := scoreData["reliability"].(float64); ok {
			score.Reliability = int(reliability)
		}
		personality.PersonalityScore = score
	}

	if behaviorPatterns, ok := data["behavior_patterns"].([]interface{}); ok {
		for _, pattern := range behaviorPatterns {
			if patternStr, ok := pattern.(string); ok {
				personality.BehaviorPatterns = append(personality.BehaviorPatterns, patternStr)
			}
		}
	}

	return personality, nil
}

// CharacterDBGroup 資料庫模型群組
type CharacterDBGroup struct {
	Character         *db.CharacterDB
	Profile           *db.CharacterProfileDB
	Localizations     []*db.CharacterLocalizationDB
	SpeechStyles      []*db.CharacterSpeechStyleDB
	Scenes            []*db.CharacterSceneDB
	States            []*db.CharacterStateDB
	EmotionalConfig   *db.CharacterEmotionalConfigDB
	NSFWConfig        *db.CharacterNSFWConfigDB
	NSFWLevels        []*db.CharacterNSFWLevelDB
	InteractionRules  []*db.CharacterInteractionRuleDB
}