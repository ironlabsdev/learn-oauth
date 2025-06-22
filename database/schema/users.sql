-- Users table - core user information
CREATE TABLE IF NOT EXISTS users
(
    id         UUID PRIMARY KEY     DEFAULT uuid_generate_v4(),
    email      TEXT UNIQUE,
    full_name  TEXT,
    avatar_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- OAuth identities table - store OAuth provider data
CREATE TABLE IF NOT EXISTS oauth_identities
(
    id           UUID PRIMARY KEY     DEFAULT uuid_generate_v4(),
    user_id      UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    provider     TEXT        NOT NULL,        -- provider 'google', 'github'
    id_token     BYTEA       NOT NULL UNIQUE, -- credential ID
    email        TEXT        NOT NULL,        -- Email from provider
    raw_data     JSONB       NOT NULL,        -- Email from provider
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users (id)
);