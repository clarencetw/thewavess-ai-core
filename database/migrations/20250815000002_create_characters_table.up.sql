-- 創建角色表
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

-- 創建索引
CREATE INDEX idx_characters_name ON characters(name);
CREATE INDEX idx_characters_type ON characters(type);
CREATE INDEX idx_characters_popularity ON characters(popularity);
CREATE INDEX idx_characters_is_active ON characters(is_active);
CREATE INDEX idx_characters_tags ON characters USING GIN(tags);
CREATE INDEX idx_characters_created_at ON characters(created_at);

-- 創建場景表
CREATE TABLE IF NOT EXISTS scenes (
    id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    character_id VARCHAR(255) NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    time_of_day VARCHAR(50) NOT NULL, -- 上午/下午/晚上
    affection_min INTEGER DEFAULT 0,  -- 最低好感度要求
    affection_max INTEGER DEFAULT 100, -- 最高好感度限制
    nsfw_level_min INTEGER DEFAULT 1, -- 最低NSFW等級
    nsfw_level_max INTEGER DEFAULT 5, -- 最高NSFW等級
    description TEXT NOT NULL,         -- 場景描述
    romantic_addition TEXT,            -- 浪漫元素附加（可選）
    weight INTEGER DEFAULT 1,         -- 權重（用於隨機選擇）
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 創建場景索引
CREATE INDEX idx_scenes_character_id ON scenes(character_id);
CREATE INDEX idx_scenes_time_of_day ON scenes(time_of_day);
CREATE INDEX idx_scenes_affection_range ON scenes(affection_min, affection_max);
CREATE INDEX idx_scenes_nsfw_level ON scenes(nsfw_level_min, nsfw_level_max);
CREATE INDEX idx_scenes_is_active ON scenes(is_active);
CREATE INDEX idx_scenes_weight ON scenes(weight);

-- 插入初始角色數據
INSERT INTO characters (id, name, type, description, avatar_url, popularity, tags, appearance, personality, background) VALUES
('char_001', '陸燁銘', 'dominant', '冷酷理性的商業鉅子，擁有強大的控制慾和佔有欲。外表俊美但內心複雜，對感情既渴望又恐懼。', 'https://placehold.co/400x400/2563eb/ffffff?text=LU', 95, ARRAY['CEO', '霸總', '冷酷'], '{"height": "185cm", "hair_color": "黑色", "eye_color": "深褐色"}', '{"traits": ["理性", "控制慾強", "佔有慾", "表面冷漠"]}', '出生於商業世家，從小接受嚴格教育，在父親的影響下成為了冷酷的商人。'),
('char_002', '沈言墨', 'ascetic', '清冷淡雅的古風美人，性格溫和但內心堅定。擅長琴棋書畫，給人一種超塵脫俗的感覺。', 'https://placehold.co/400x400/10b981/ffffff?text=SHEN', 88, ARRAY['古風', '文雅', '清冷'], '{"height": "175cm", "hair_color": "墨黑", "eye_color": "清澈黑眸"}', '{"traits": ["溫和", "堅定", "淡雅", "超然"]}', '出身書香門第，自幼熟讀詩書，性格淡然，對世俗名利看得很淡。');

-- 插入陸燁銘的場景數據
INSERT INTO scenes (character_id, time_of_day, affection_min, affection_max, nsfw_level_min, nsfw_level_max, description, romantic_addition, weight) VALUES
-- 上午場景
('char_001', '上午', 0, 100, 1, 5, '陽光透過辦公室的百葉窗灑在陸寒淵的側臉上，他專注地處理文件的樣子格外迷人', '，空氣中似乎都瀰漫著曖昧的氣息', 3),
('char_001', '上午', 0, 100, 1, 5, '辦公室裡瀰漫著淡淡的咖啡香，陸寒淵抬頭看向你時，眼中閃爍著溫柔的光芒', '，你們之間的距離越來越近', 2),
-- 下午場景  
('char_001', '下午', 0, 100, 1, 5, '下午的陽光將辦公室染成金黃色，陸寒淵放下手中的筆，深邃的眼眸注視著你', '，他的呼吸變得有些急促', 3),
('char_001', '下午', 0, 100, 1, 5, '會議室裡只剩下你們兩人，夕陽西下，陸寒淵的輪廓在光影中顯得格外性感', '，房間裡的溫度似乎在上升', 2),
-- 晚上場景
('char_001', '晚上', 0, 100, 1, 5, '夜色籠罩著城市，辦公室裡燈光昏暗，陸寒淵緩緩起身走向你', '，空氣中似乎都瀰漫著曖昧的氣息', 3),
('char_001', '晚上', 0, 100, 1, 5, '城市的霓虹透過落地窗映照在陸寒淵的臉上，他的眼神變得更加深邃迷人', '，你們之間的距離越來越近', 2);

-- 插入沈言墨的場景數據
INSERT INTO scenes (character_id, time_of_day, affection_min, affection_max, nsfw_level_min, nsfw_level_max, description, romantic_addition, weight) VALUES
-- 上午場景
('char_002', '上午', 0, 100, 1, 5, '醫院的晨光透過窗戶灑進診療室，沈言墨溫和地整理著醫療器械', '，他溫柔的目光讓人心動', 3),
('char_002', '上午', 0, 100, 1, 5, '白大褂在晨光中顯得格外潔白，沈言墨溫柔的笑容如春風般溫暖', '，空氣中瀰漫著淡淡的藥香和他的體香', 2),
-- 下午場景
('char_002', '下午', 0, 100, 1, 5, '午後的陽光讓診療室變得溫馨，沈言墨摘下聽診器，專注地看著你', '，他的關懷讓人感到無比安心', 3),
('char_002', '下午', 0, 100, 1, 5, '醫院的走廊裡人來人往，但沈言墨的注意力完全在你身上', '，你們之間似乎只有彼此', 2),
-- 晚上場景
('char_002', '晚上', 0, 100, 1, 5, '夜班的醫院格外安靜，值班室裡只有你和沈言墨，氛圍變得親密而溫馨', '，月光下的他格外動人', 3),
('char_002', '晚上', 0, 100, 1, 5, '月光透過窗戶灑在沈言墨的白大褂上，他疲憊卻溫柔的笑容讓人心動', '，夜晚的靜謐讓你們的心更加靠近', 2);