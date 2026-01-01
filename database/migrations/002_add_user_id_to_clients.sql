DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'clients' AND column_name = 'user_id'
    ) THEN
        ALTER TABLE clients ADD COLUMN user_id BIGINT;
    END IF;
END $$;

DELETE FROM clients WHERE user_id IS NULL;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'clients' 
        AND column_name = 'user_id' 
        AND is_nullable = 'YES'
    ) THEN
        ALTER TABLE clients ALTER COLUMN user_id SET NOT NULL;
        ALTER TABLE clients ALTER COLUMN user_id SET DEFAULT 0;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_clients_user_id ON clients(user_id);
