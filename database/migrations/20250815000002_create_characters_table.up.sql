-- Create characters table
CREATE TABLE IF NOT EXISTS characters (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    description TEXT,
    avatar_url VARCHAR(512),
    popularity INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    tags TEXT[] DEFAULT '{}',
    appearance JSONB DEFAULT '{}',
    personality JSONB DEFAULT '{}',
    background TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_characters_name ON characters(name);
CREATE INDEX idx_characters_type ON characters(type);
CREATE INDEX idx_characters_popularity ON characters(popularity);
CREATE INDEX idx_characters_is_active ON characters(is_active);
CREATE INDEX idx_characters_tags ON characters USING GIN(tags);
CREATE INDEX idx_characters_created_at ON characters(created_at);

-- Insert initial character data
INSERT INTO characters (id, name, type, description, avatar_url, popularity, tags, appearance, personality, background) VALUES
('char_001', '陸燁銘', 'dominant', '冷酷理性的商業鉅子，擁有強大的控制慾和佔有欲。外表俊美但內心複雜，對感情既渴望又恐懼。', 'https://placehold.co/400x400/2563eb/ffffff?text=LU', 95, ARRAY['CEO', '霸總', '冷酷'], '{"height": "185cm", "hair_color": "黑色", "eye_color": "深褐色"}', '{"traits": ["理性", "控制慾強", "佔有慾", "表面冷漠"]}', '出生於商業世家，從小接受嚴格教育，在父親的影響下成為了冷酷的商人。'),
('char_002', '沈言墨', 'ascetic', '清冷淡雅的古風美人，性格溫和但內心堅定。擅長琴棋書畫，給人一種超塵脫俗的感覺。', 'https://placehold.co/400x400/10b981/ffffff?text=SHEN', 88, ARRAY['古風', '文雅', '清冷'], '{"height": "175cm", "hair_color": "墨黑", "eye_color": "清澈黑眸"}', '{"traits": ["溫和", "堅定", "淡雅", "超然"]}', '出身書香門第，自幼熟讀詩書，性格淡然，對世俗名利看得很淡。');