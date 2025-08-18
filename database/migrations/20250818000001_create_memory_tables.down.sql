-- 刪除記憶相關表（按依賴順序）
DROP INDEX IF EXISTS idx_memory_personal_info_memory_id;
DROP INDEX IF EXISTS idx_memory_dislikes_memory_id;
DROP INDEX IF EXISTS idx_memory_milestones_memory_id;
DROP INDEX IF EXISTS idx_memory_nicknames_memory_id;
DROP INDEX IF EXISTS idx_memory_preferences_memory_id;
DROP INDEX IF EXISTS idx_long_term_memories_user_character;

DROP TABLE IF EXISTS memory_personal_info;
DROP TABLE IF EXISTS memory_dislikes;
DROP TABLE IF EXISTS memory_milestones;
DROP TABLE IF EXISTS memory_nicknames;
DROP TABLE IF EXISTS memory_preferences;
DROP TABLE IF EXISTS long_term_memories;