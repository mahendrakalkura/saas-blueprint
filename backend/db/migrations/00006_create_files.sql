-- +goose Up
-- +goose StatementBegin
CREATE TABLE files (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    filename VARCHAR(500) NOT NULL,
    original_filename VARCHAR(500) NOT NULL,
    mime_type VARCHAR(255) NOT NULL,
    size_bytes BIGINT NOT NULL,
    storage_key VARCHAR(1000) NOT NULL UNIQUE,
    storage_bucket VARCHAR(255) NOT NULL,
    is_public BOOLEAN DEFAULT FALSE,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_files_user_id ON files(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_files_organization_id ON files(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_files_storage_key ON files(storage_key);
CREATE INDEX idx_files_created_at ON files(created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS files;
-- +goose StatementEnd
