-- name: CreateOrder :one
INSERT INTO orders (user_id, number, status)
VALUES (@user_id, @number, @status)
RETURNING id, user_id, number, status, uploaded_at;

-- name: GetOrdersByUser :many
SELECT id, user_id, number, status, uploaded_at
FROM orders
WHERE user_id = @user_id
ORDER BY uploaded_at DESC;

-- name: GetOrderByNumber :one
SELECT id, user_id, number, status, uploaded_at
FROM orders
WHERE number = @number;
