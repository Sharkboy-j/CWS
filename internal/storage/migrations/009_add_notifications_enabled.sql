ALTER TABLE user_states
ADD COLUMN IF NOT EXISTS notifications_enabled BOOLEAN DEFAULT TRUE;

UPDATE user_states SET notifications_enabled = TRUE WHERE notifications_enabled IS NULL;

