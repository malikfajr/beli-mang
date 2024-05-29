CREATE TABLE IF NOT EXISTS merchants(
    id CHAR(26) PRIMARY KEY,
    username_admin VARCHAR(30) nOT NULL,
    name VARCHAR(30) NOT NULL,
    category VARCHAR(21) NOT NULL,
    image_url TEXT NOT NULL,
    lat NUMERIC NOT NULL,
    long NUMERIC NOT NULL,
    geohash VARCHAR(12) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),

    FOREIGN KEY (username_admin) REFERENCES users(username) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_merchant_name ON merchants(LOWER(name));

CREATE INDEX IF NOT EXISTS idx_merchant_category ON merchants(category);

CREATE INDEX IF NOT EXISTS idx_merchant_created_at ON merchants(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_merchant_geohash ON merchants(geohash);