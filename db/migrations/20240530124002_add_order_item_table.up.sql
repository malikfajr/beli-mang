CREATE TABLE IF NOT EXISTS order_items(
    id BIGSERIAL PRIMARY KEY,
    order_id CHAR(26),
    merchant_id CHAR(26),
    item_id CHAR(26),
    quantity INT,
    created_at TIMESTAMPTZ DEFAULT NOW(),

    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (merchant_id) REFERENCES merchants(id) ON UPDATE CASCADE ON DELETE CASCADE,
    FOREIGN KEY (item_id) REFERENCES products(id) ON UPDATE CASCADE ON DELETE CASCADE
)