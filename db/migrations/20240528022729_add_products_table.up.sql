CREATE TABLE IF NOT EXISTS products(
    id CHAR(26) PRIMARY KEY,
    merchant_id CHAR(26) NOT NULL,
    name VARCHAR(30) NOT NULL,
    category VARCHAR(10) NOT NULL,
    price INT NOT NULL,
    image_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    FOREIGN KEY (merchant_id) REFERENCES merchants(id) 
    ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_product_name on products(LOWER(name));

CREATE INDEX IF NOT EXISTS idx_product_created_at ON products(created_at DESC);