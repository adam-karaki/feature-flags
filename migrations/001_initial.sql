CREATE TABLE IF NOT EXISTS flags (
    name TEXT PRIMARY KEY,
    enabled BOOLEAN NOT NULL,
    rules JSONB NOT NULL,
    version BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS flags_updated_at_idx ON flags(updated_at);
