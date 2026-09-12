CREATE TABLE seigen_policies (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    identifier TEXT NOT NULL UNIQUE,
    limiter_config JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


CREATE TABLE seigen_policy_groups (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    limiter_config JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


CREATE TABLE seigen_policy_group_members (
    group_id BIGINT NOT NULL REFERENCES seigen_policy_groups(id) ON DELETE CASCADE,
    identifier TEXT NOT NULL UNIQUE
);

CREATE INDEX idx_seigen_policy_group_members_group_id
ON seigen_policy_group_members(group_id);