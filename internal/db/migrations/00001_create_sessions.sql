-- +goose Up
-- +goose StatementBegin
CREATE TABLE sessions (
    id            TEXT PRIMARY KEY,
    user_id       TEXT        NOT NULL,
    user_email    TEXT        NOT NULL DEFAULT '',
    user_name     TEXT        NOT NULL DEFAULT '',
    claims        JSONB       NOT NULL DEFAULT '{}'::jsonb,
    refresh_token TEXT        NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at    TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_sessions_user_id ON sessions (user_id);
CREATE INDEX idx_sessions_expires_at ON sessions (expires_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS sessions;
-- +goose StatementEnd
