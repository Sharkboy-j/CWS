-- Add recommended_torrents_count to user_states
ALTER TABLE user_states
ADD COLUMN IF NOT EXISTS recommended_torrents_count integer DEFAULT 3;

-- Ensure existing rows have default value
UPDATE user_states SET recommended_torrents_count = 3 WHERE recommended_torrents_count IS NULL;

