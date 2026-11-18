-- name: CreateUser :one
INSERT INTO users (login, password_hash)
VALUES (@login, @password_hash)
RETURNING id, login;


-- name: GetUserByLogin :one
SELECT id, login, password_hash
FROM users
WHERE login = @login;
