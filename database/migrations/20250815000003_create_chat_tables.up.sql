-- Create chat_sessions table
CREATE TABLE IF NOT EXISTS chat_sessions (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    character_id VARCHAR(255) NOT NULL,
    title VARCHAR(255),
    mode VARCHAR(50) DEFAULT 'normal',
    status VARCHAR(50) DEFAULT 'active',
    tags TEXT[] DEFAULT '{}',
    message_count INTEGER DEFAULT 0,
    total_characters INTEGER DEFAULT 0,
    last_message_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (character_id) REFERENCES characters(id) ON DELETE CASCADE
);

-- Create messages table
CREATE TABLE IF NOT EXISTS messages (
    id VARCHAR(255) PRIMARY KEY,
    session_id VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    content TEXT NOT NULL,
    scene_description TEXT,
    character_action TEXT,
    emotional_state JSONB DEFAULT '{}',
    ai_engine VARCHAR(100),
    response_time_ms INTEGER,
    nsfw_level INTEGER DEFAULT 0,
    is_regenerated BOOLEAN DEFAULT FALSE,
    regeneration_reason TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (session_id) REFERENCES chat_sessions(id) ON DELETE CASCADE
);

-- Create user_sessions table (for multi-user sessions in the future)
CREATE TABLE IF NOT EXISTS user_sessions (
    user_id VARCHAR(255),
    session_id VARCHAR(255),
    role VARCHAR(50) DEFAULT 'owner',
    joined_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (user_id, session_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (session_id) REFERENCES chat_sessions(id) ON DELETE CASCADE
);

-- Create indexes for chat_sessions
CREATE INDEX idx_chat_sessions_user_id ON chat_sessions(user_id);
CREATE INDEX idx_chat_sessions_character_id ON chat_sessions(character_id);
CREATE INDEX idx_chat_sessions_status ON chat_sessions(status);
CREATE INDEX idx_chat_sessions_mode ON chat_sessions(mode);
CREATE INDEX idx_chat_sessions_updated_at ON chat_sessions(updated_at);
CREATE INDEX idx_chat_sessions_tags ON chat_sessions USING GIN(tags);

-- Create indexes for messages
CREATE INDEX idx_messages_session_id ON messages(session_id);
CREATE INDEX idx_messages_role ON messages(role);
CREATE INDEX idx_messages_created_at ON messages(created_at);
CREATE INDEX idx_messages_nsfw_level ON messages(nsfw_level);

-- Create indexes for user_sessions
CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_session_id ON user_sessions(session_id);