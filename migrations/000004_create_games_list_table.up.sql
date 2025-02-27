CREATE TABLE IF NOT EXISTS games_list (
    -- base fields
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version INTEGER NOT NULL DEFAULT 1,
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    -- fields specific to games
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    logo TEXT NOT NULL,
    src TEXT NOT NULL,
    controls TEXT NOT NULL,
    has_score BOOLEAN NOT NULL DEFAULT false
);

