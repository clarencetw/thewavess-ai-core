-- 創建情感狀態表，用於持久化情感數據
CREATE TABLE IF NOT EXISTS emotion_states (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    character_id VARCHAR(255) NOT NULL,
    affection INTEGER NOT NULL DEFAULT 30,
    mood VARCHAR(50) NOT NULL DEFAULT 'neutral',
    relationship VARCHAR(50) NOT NULL DEFAULT 'stranger',
    intimacy_level VARCHAR(50) NOT NULL DEFAULT 'distant',
    total_interactions INTEGER DEFAULT 0,
    last_interaction TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (character_id) REFERENCES characters(id) ON DELETE CASCADE,
    UNIQUE (user_id, character_id)
);

-- 創建情感歷史表，用於追蹤情感變化
CREATE TABLE IF NOT EXISTS emotion_history (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    character_id VARCHAR(255) NOT NULL,
    old_affection INTEGER NOT NULL,
    new_affection INTEGER NOT NULL,
    affection_change INTEGER NOT NULL,
    old_mood VARCHAR(50) NOT NULL,
    new_mood VARCHAR(50) NOT NULL,
    trigger_type VARCHAR(100) NOT NULL,
    trigger_content TEXT,
    context JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (character_id) REFERENCES characters(id) ON DELETE CASCADE
);

-- 創建情感里程碑表，用於記錄關係里程碑
CREATE TABLE IF NOT EXISTS emotion_milestones (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    character_id VARCHAR(255) NOT NULL,
    milestone_type VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    affection_level INTEGER NOT NULL,
    achieved_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (character_id) REFERENCES characters(id) ON DELETE CASCADE
);

-- 為情感狀態表創建索引
CREATE INDEX idx_emotion_states_user_id ON emotion_states(user_id);
CREATE INDEX idx_emotion_states_character_id ON emotion_states(character_id);
CREATE INDEX idx_emotion_states_affection ON emotion_states(affection);
CREATE INDEX idx_emotion_states_updated_at ON emotion_states(updated_at);

-- 為情感歷史表創建索引
CREATE INDEX idx_emotion_history_user_id ON emotion_history(user_id);
CREATE INDEX idx_emotion_history_character_id ON emotion_history(character_id);
CREATE INDEX idx_emotion_history_created_at ON emotion_history(created_at);
CREATE INDEX idx_emotion_history_trigger_type ON emotion_history(trigger_type);

-- 為情感里程碑表創建索引  
CREATE INDEX idx_emotion_milestones_user_id ON emotion_milestones(user_id);
CREATE INDEX idx_emotion_milestones_character_id ON emotion_milestones(character_id);
CREATE INDEX idx_emotion_milestones_achieved_at ON emotion_milestones(achieved_at);