-- name: CreatePerson :one
INSERT INTO persons (name, created_at, updated_at)
VALUES ($1, now(), now())
RETURNING id, name, created_at, updated_at;

-- name: ListPersons :many
SELECT id, name, created_at, updated_at
FROM persons
ORDER BY id;
