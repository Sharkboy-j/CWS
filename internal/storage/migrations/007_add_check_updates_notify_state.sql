ALTER TABLE user_states
ADD COLUMN IF NOT EXISTS check_updates_notify_message_id INTEGER DEFAULT NULL;

ALTER TABLE user_states
ADD COLUMN IF NOT EXISTS check_updates_notify_payload_hash TEXT DEFAULT NULL;

ALTER TABLE user_states
ADD COLUMN IF NOT EXISTS check_updates_notify_missing_hashes TEXT DEFAULT NULL;

