CREATE EXTENSION IF NOT EXISTS CITEXT;
CREATE TABLE IF NOT EXISTS auth_users (
    -- base fields
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version INTEGER NOT NULL DEFAULT 1,
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    -- fields specific to users
    email CITEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    profile_picture TEXT, -- nullable. server handles default image if no profile_picture
    password_hash BYTEA,  -- nullable. allow logins via OAuth
    provider TEXT,        -- nullable. allow logins via email/password
    role_id SMALLINT DEFAULT 1 NOT NULL REFERENCES auth_roles ON DELETE SET DEFAULT,
    is_verified BOOLEAN NOT NULL DEFAULT FALSE
);
