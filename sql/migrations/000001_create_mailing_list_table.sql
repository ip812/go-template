-- +goose Up
CREATE TABLE IF NOT EXISTS mailing_list (
    id bigserial PRIMARY KEY,
    email text NOT NULL UNIQUE
);

-- +goose Down
DROP TABLE mailing_list;
