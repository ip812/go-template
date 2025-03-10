-- +goose Up
CREATE TABLE IF NOT EXISTS mailing_list (
    id bigserial PRIMARY KEY,
    email text NOT NULL UNIQUE,
    last_sent_at timestamp,
    created_at timestamp NOT NULL
);

-- +goose Down
DROP TABLE mailing_list;

