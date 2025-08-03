CREATE TABLE IF NOT EXISTS products
(
    id BIGSERIAL PRIMARY KEY,
    label VARCHAR(50),
    description VARCHAR(200),
    price INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS product_history (
    id BIGSERIAL PRIMARY KEY,
    product_id INT NOT NULL,
    label VARCHAR(50),
    description VARCHAR(200),
    price INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);