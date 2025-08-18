-- 創建長期記憶表
CREATE TABLE IF NOT EXISTS long_term_memories (
    id VARCHAR(20) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    user_id VARCHAR(20) NOT NULL,
    character_id VARCHAR(20) NOT NULL,
    last_updated TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, character_id)
);

-- 創建偏好表
CREATE TABLE IF NOT EXISTS memory_preferences (
    id VARCHAR(20) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    memory_id VARCHAR(20) NOT NULL REFERENCES long_term_memories(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    category VARCHAR(50) NOT NULL,
    importance INTEGER NOT NULL DEFAULT 5,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 創建稱呼表
CREATE TABLE IF NOT EXISTS memory_nicknames (
    id VARCHAR(20) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    memory_id VARCHAR(20) NOT NULL REFERENCES long_term_memories(id) ON DELETE CASCADE,
    nickname VARCHAR(100) NOT NULL,
    frequency INTEGER NOT NULL DEFAULT 1,
    last_used TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 創建里程碑表
CREATE TABLE IF NOT EXISTS memory_milestones (
    id VARCHAR(20) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    memory_id VARCHAR(20) NOT NULL REFERENCES long_term_memories(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    description TEXT NOT NULL,
    affection INTEGER NOT NULL,
    date TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 創建禁忌表
CREATE TABLE IF NOT EXISTS memory_dislikes (
    id VARCHAR(20) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    memory_id VARCHAR(20) NOT NULL REFERENCES long_term_memories(id) ON DELETE CASCADE,
    topic TEXT NOT NULL,
    severity INTEGER NOT NULL DEFAULT 3,
    evidence TEXT,
    recorded_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 創建個人信息表
CREATE TABLE IF NOT EXISTS memory_personal_info (
    id VARCHAR(20) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    memory_id VARCHAR(20) NOT NULL REFERENCES long_term_memories(id) ON DELETE CASCADE,
    info_type VARCHAR(50) NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(memory_id, info_type)
);

-- 創建索引
CREATE INDEX IF NOT EXISTS idx_long_term_memories_user_character ON long_term_memories(user_id, character_id);
CREATE INDEX IF NOT EXISTS idx_memory_preferences_memory_id ON memory_preferences(memory_id);
CREATE INDEX IF NOT EXISTS idx_memory_nicknames_memory_id ON memory_nicknames(memory_id);
CREATE INDEX IF NOT EXISTS idx_memory_milestones_memory_id ON memory_milestones(memory_id);
CREATE INDEX IF NOT EXISTS idx_memory_dislikes_memory_id ON memory_dislikes(memory_id);
CREATE INDEX IF NOT EXISTS idx_memory_personal_info_memory_id ON memory_personal_info(memory_id);