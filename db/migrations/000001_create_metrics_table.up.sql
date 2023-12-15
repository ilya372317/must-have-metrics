BEGIN;
CREATE TYPE metric_type AS ENUM ('gauge', 'counter');
CREATE TABLE IF NOT EXISTS metrics
(
    id VARCHAR(300) NOT NULL PRIMARY KEY,
    "type"      metric_type,
    float_value double precision,
    int_value   BIGINT
)