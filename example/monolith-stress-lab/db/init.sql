CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL
);

DO $$
DECLARE
    i INT := 1;
    p_hash VARCHAR(255) := '$2a$10$zt7FHwiiK36eg8B7gU442eLTJVtWF4d0a3svAeqfkRFT1zt5TREaW'; 
BEGIN
    WHILE i <= 10000 LOOP
        INSERT INTO users (username, password_hash) VALUES ('user' || i, p_hash) ON CONFLICT DO NOTHING;
        i := i + 1;
    END LOOP;
END $$;
