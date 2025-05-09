CREATE TABLE IF NOT EXISTS products
(
    id
    BIGSERIAL
    PRIMARY
    KEY,
    label
    VARCHAR(50),
    description
    VARCHAR(200),
    price
    INT,
    created_at
    TIMESTAMP
    DEFAULT
    CURRENT_TIMESTAMP
);