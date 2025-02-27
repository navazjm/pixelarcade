
CREATE TABLE IF NOT EXISTS games_scores (
    -- base fields
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version INTEGER NOT NULL DEFAULT 1,
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    -- fields specific to scores
    game_id BIGINT NOT NULL REFERENCES games_list ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES auth_users ON DELETE CASCADE,
    score BIGINT NOT NULL
);
