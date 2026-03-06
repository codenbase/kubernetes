CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS articles (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    size_type VARCHAR(20) NOT NULL
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

    -- Insert 100 small articles (~100 Bytes payload)
    i := 1;
    WHILE i <= 100 LOOP
        INSERT INTO articles (id, title, content, size_type) 
        VALUES (i, 'Small Article ' || i, 'This is a short summary content for testing small packet DB reads. ID:' || i, 'small') 
        ON CONFLICT DO NOTHING;
        i := i + 1;
    END LOOP;

    -- Insert 100 large articles (~50KB payload)
    WHILE i <= 200 LOOP
        INSERT INTO articles (id, title, content, size_type) 
        VALUES (i, 'Large Article ' || i, repeat('A', 50000) || ' ID:' || i, 'large') 
        ON CONFLICT DO NOTHING;
        i := i + 1;
    END LOOP;
END $$;
