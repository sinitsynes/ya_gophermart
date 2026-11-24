-- name: CreateOrder :one
INSERT INTO orders (user_id, number, status)
VALUES (@user_id, @number, @status)
RETURNING id, user_id, number, status, uploaded_at;

-- name: GetOrdersByUser :many
SELECT id, user_id, number, status, accrual, uploaded_at
FROM orders
WHERE user_id = @user_id
ORDER BY uploaded_at DESC;

-- name: GetOrderByNumber :one
SELECT id, user_id, number, status, accrual, uploaded_at
FROM orders
WHERE number = @number;

-- name: UpdateOrdersAccrual :exec
UPDATE orders AS o
SET
    accrual = CASE WHEN upd.has_accrual THEN upd.accrual ELSE NULL END,
    status = upd.status
FROM (
    SELECT
        UNNEST(@numbers::text[]) AS number,
        UNNEST(@statuses::int[]) AS status,
        UNNEST(@accruals::real[]) AS accrual,
        UNNEST(@has_accrual::boolean[]) AS has_accrual
) AS upd
WHERE o.number = upd.number;
