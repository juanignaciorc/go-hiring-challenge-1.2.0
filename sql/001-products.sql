CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    code VARCHAR(32),
    price DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Useful indexes on products
CREATE INDEX IF NOT EXISTS idx_products_code ON products (code);
CREATE INDEX IF NOT EXISTS idx_products_price ON products (price);
CREATE INDEX IF NOT EXISTS idx_products_created_at ON products (created_at);
CREATE INDEX IF NOT EXISTS idx_products_updated_at ON products (updated_at);

-- Make code unique if it's meant to be a unique product identifier
-- Uncomment the following line if you need code to be unique:
-- CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx_products_code_unique ON products(code);