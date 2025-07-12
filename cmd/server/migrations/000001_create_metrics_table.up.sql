CREATE TABLE metrics (
    id VARCHAR(255) PRIMARY KEY,
    type VARCHAR(255) NOT NULL,
    delta int NULL,
    value float8 NULL,
    hash VARCHAR(255) NULL
)