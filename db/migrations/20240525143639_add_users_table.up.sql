CREATE TABLE IF NOT EXISTS users(
    username VARCHAR(30) PRIMARY KEY,
    password CHAR(60) NOT NULL,
    email VARCHAR(100) NOT NULL,
    admin BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_email ON users(email);