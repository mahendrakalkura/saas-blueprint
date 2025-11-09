-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
ADD COLUMN mfa_enabled BOOLEAN DEFAULT FALSE,
ADD COLUMN mfa_secret VARCHAR(255),
ADD COLUMN mfa_backup_codes TEXT[];

CREATE INDEX idx_users_mfa_enabled ON users(mfa_enabled);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users
DROP COLUMN IF EXISTS mfa_enabled,
DROP COLUMN IF EXISTS mfa_secret,
DROP COLUMN IF EXISTS mfa_backup_codes;

DROP INDEX IF EXISTS idx_users_mfa_enabled;
-- +goose StatementEnd
