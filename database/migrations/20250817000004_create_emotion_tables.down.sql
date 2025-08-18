-- Drop emotion tables in reverse order
DROP INDEX IF EXISTS idx_emotion_milestones_achieved_at;
DROP INDEX IF EXISTS idx_emotion_milestones_character_id;
DROP INDEX IF EXISTS idx_emotion_milestones_user_id;

DROP INDEX IF EXISTS idx_emotion_history_trigger_type;
DROP INDEX IF EXISTS idx_emotion_history_created_at;
DROP INDEX IF EXISTS idx_emotion_history_character_id;
DROP INDEX IF EXISTS idx_emotion_history_user_id;

DROP INDEX IF EXISTS idx_emotion_states_updated_at;
DROP INDEX IF EXISTS idx_emotion_states_affection;
DROP INDEX IF EXISTS idx_emotion_states_character_id;
DROP INDEX IF EXISTS idx_emotion_states_user_id;

DROP TABLE IF EXISTS emotion_milestones;
DROP TABLE IF EXISTS emotion_history;
DROP TABLE IF EXISTS emotion_states;