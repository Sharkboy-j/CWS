ALTER TABLE user_states
ADD COLUMN IF NOT EXISTS notify_bot_subscribed BOOLEAN DEFAULT FALSE;

UPDATE user_states SET notify_bot_subscribed = FALSE WHERE notify_bot_subscribed IS NULL;

