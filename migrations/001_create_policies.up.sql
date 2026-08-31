CREATE TABLE policies {
    id UUID PRIMARY KEY,
    identifier TEXT NOT NULL UNIQUE,
    limiter_config JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
};