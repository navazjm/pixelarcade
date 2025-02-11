CREATE TABLE IF NOT EXISTS auth_roles (
    -- base fields
    id SMALLSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version INTEGER NOT NULL DEFAULT 1,
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    -- fields specific to roles
    name TEXT UNIQUE NOT NULL
);

INSERT INTO auth_roles (id, name, is_active) VALUES (1, 'Basic', TRUE);
INSERT INTO auth_roles (id, name, is_active) VALUES (2, 'Admin', TRUE);

-- Needed to update Postgres autoincrement value
SELECT setval(pg_get_serial_sequence('auth_roles', 'id'), (SELECT MAX(id) FROM auth_roles));
