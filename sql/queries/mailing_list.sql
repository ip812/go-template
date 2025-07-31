-- name: AddEmailToMailingList :one
INSERT INTO mailing_list (id, email)
    VALUES ($1, $2)
RETURNING
    id, email;
