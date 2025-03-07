-- +goose Up
CREATE TABLE refresh_tokens (
    token TEXT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP DEFAULT NULL
);

CREATE INDEX refresh_tokens_user_id_idx ON refresh_tokens(user_id);

CREATE INDEX refresh_tokens_expires_at_idx ON refresh_tokens(expires_at);

-- +goose Down
DROP TABLE refresh_tokens;