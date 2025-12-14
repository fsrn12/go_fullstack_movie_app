-- +goose Up
-- +goose StatementBegin
CREATE TABLE tokens (
    id SERIAL PRIMARY KEY,
    hash BYTEA NOT NULL,
    user_id INT NOT NULL,
    expiry TIMESTAMP NOT NULL,
    scope VARCHAR(255) NOT NULL,
    CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES users(id)
        ON DELETE CASCADE         -- Ensures that if a user is deleted, their tokens are also deleted
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tokens;
-- +goose StatementEnd
