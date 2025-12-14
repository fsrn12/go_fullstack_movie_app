-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_tokens_hash ON tokens (hash);
CREATE INDEX idx_tokens_hash_scope ON tokens (hash, scope);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
DROP INDEX IF EXISTS idx_tokens_hash_scope;
DROP INDEX IF EXISTS idx_tokens_hash;
