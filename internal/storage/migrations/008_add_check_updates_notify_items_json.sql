ALTER TABLE user_states
ADD COLUMN IF NOT EXISTS check_updates_notify_items_json TEXT DEFAULT NULL;

