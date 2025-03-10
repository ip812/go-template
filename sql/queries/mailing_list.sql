-- name: AddEmailToMailingList :one
INSERT INTO mailing_list (id, email, last_sent_at, created_at)
    VALUES ($1, $2, NULL, $3)
RETURNING
    id, email, last_sent_at, created_at;

