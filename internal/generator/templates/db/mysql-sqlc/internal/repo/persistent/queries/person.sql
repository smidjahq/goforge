-- name: CreatePerson :execresult
INSERT INTO persons (name, created_at, updated_at)
VALUES (?, NOW(), NOW());

-- name: ListPersons :many
SELECT id, name, created_at, updated_at FROM persons ORDER BY id;
