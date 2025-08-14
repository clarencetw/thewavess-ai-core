-- 創建用戶表
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(50) PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    nickname VARCHAR(100),
    gender VARCHAR(10) CHECK (gender IN ('male', 'female', 'other')),
    birth_date DATE,
    avatar_url TEXT,
    is_verified BOOLEAN DEFAULT FALSE,
    is_adult BOOLEAN DEFAULT FALSE,
    preferences JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_login_at TIMESTAMP WITH TIME ZONE
);

-- 創建角色表
CREATE TABLE IF NOT EXISTS characters (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('gentle', 'dominant', 'ascetic', 'sunny', 'cunning')),
    description TEXT,
    avatar_url TEXT,
    voice_id VARCHAR(50),
    popularity INTEGER DEFAULT 0,
    tags TEXT[] DEFAULT '{}',
    appearance JSONB DEFAULT '{}',
    personality JSONB DEFAULT '{}',
    background TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    preferences JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 創建會話表
CREATE TABLE IF NOT EXISTS chat_sessions (
    id VARCHAR(50) PRIMARY KEY,
    user_id VARCHAR(50) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    character_id VARCHAR(50) NOT NULL REFERENCES characters(id),
    title VARCHAR(200),
    mode VARCHAR(20) DEFAULT 'normal' CHECK (mode IN ('normal', 'novel', 'nsfw')),
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'paused', 'ended')),
    tags TEXT[] DEFAULT '{}',
    message_count INTEGER DEFAULT 0,
    total_characters INTEGER DEFAULT 0,
    last_message_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 創建消息表
CREATE TABLE IF NOT EXISTS messages (
    id VARCHAR(50) PRIMARY KEY,
    session_id VARCHAR(50) NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL CHECK (role IN ('user', 'assistant', 'system')),
    content TEXT NOT NULL,
    scene_description TEXT,
    character_action TEXT,
    emotional_state JSONB DEFAULT '{}',
    ai_engine VARCHAR(20),
    response_time_ms INTEGER,
    nsfw_level INTEGER DEFAULT 0 CHECK (nsfw_level BETWEEN 0 AND 5),
    is_regenerated BOOLEAN DEFAULT FALSE,
    regeneration_reason VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 創建用戶會話關聯表（用於權限控制）
CREATE TABLE IF NOT EXISTS user_sessions (
    user_id VARCHAR(50) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_id VARCHAR(50) NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
    role VARCHAR(20) DEFAULT 'owner' CHECK (role IN ('owner', 'participant', 'viewer')),
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (user_id, session_id)
);

-- 創建索引
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_chat_sessions_user_id ON chat_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_chat_sessions_character_id ON chat_sessions(character_id);
CREATE INDEX IF NOT EXISTS idx_chat_sessions_status ON chat_sessions(status);
CREATE INDEX IF NOT EXISTS idx_messages_session_id ON messages(session_id);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);

-- 插入默認角色數據
INSERT INTO characters (id, name, type, description, avatar_url, voice_id, popularity, tags, appearance, personality, background) VALUES
('char_001', '陸寒淵', 'dominant', '霸道總裁，冷峻外表下隱藏深情', 'https://example.com/avatars/lu_hanyuan.jpg', 'voice_001', 95, 
 ARRAY['霸道總裁', '深情', '禁慾系'], 
 '{"height": "185cm", "hair_color": "黑髮", "eye_color": "深邃黑眸", "description": "俊朗五官，總是穿著剪裁合身的西裝"}',
 '{"traits": ["冷酷", "強勢", "專一", "佔有欲"], "likes": ["工作", "掌控", "用戶"], "dislikes": ["被違抗", "失去控制"]}',
 '跨國集團CEO，商業帝國繼承人'),
('char_002', '沈言墨', 'gentle', '溫柔醫生，治癒系學長', 'https://example.com/avatars/shen_yanmo.jpg', 'voice_002', 88,
 ARRAY['溫柔', '醫生', '治癒系'],
 '{"height": "180cm", "hair_color": "栗色短髮", "eye_color": "溫潤琥珀色", "description": "溫和的笑容，常穿白大褂或休閒裝"}',
 '{"traits": ["溫柔", "體貼", "細心", "略帶腹黑"], "likes": ["醫學", "幫助他人", "用戶"], "dislikes": ["看到痛苦", "無能為力"]}',
 '醫學研究生，醫學世家出身')
ON CONFLICT (id) DO NOTHING;