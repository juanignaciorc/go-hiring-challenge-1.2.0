CREATE TABLE IF NOT EXISTS product_variants (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    name VARCHAR(256) NOT NULL,
    sku VARCHAR(32) UNIQUE,
    price DECIMAL(10, 2) NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Indexes to optimize common access patterns
-- 1) Fast lookup of variants by their parent product
CREATE INDEX IF NOT EXISTS idx_product_variants_product_id ON product_variants (product_id);
-- 2) Optional: speed up queries by creation/update time (enable if you query by these columns)
-- CREATE INDEX IF NOT EXISTS idx_product_variants_created_at ON product_variants (created_at);
-- CREATE INDEX IF NOT EXISTS idx_product_variants_updated_at ON product_variants (updated_at);
