-- name: CreateOrder :one
INSERT INTO orders (user_id, number, status)
VALUES (@user_id, @number, @status)
RETURNING id, user_id, number, status, uploaded_at;

-- name: CreateOrderAccrual :exec
INSERT INTO order_accruals (order_id)
VALUES (@order_id);

-- name: GetOrdersByUser :many
SELECT o.id, o.user_id, o.number, o.status, oa.accrual, o.uploaded_at
FROM orders o
LEFT JOIN order_accruals oa ON o.id = oa.order_id
WHERE o.user_id = @user_id
ORDER BY o.uploaded_at DESC;

-- name: GetOrderByNumber :one
SELECT o.id, o.user_id, o.number, o.status, oa.accrual, o.uploaded_at
FROM orders o
LEFT JOIN order_accruals oa ON o.id = oa.order_id
WHERE o.number = @number;

-- name: UpdateOrderStatuses :exec
UPDATE orders AS o
SET status = upd.status
FROM (
    SELECT
        UNNEST(@ids::int[]) AS id,
        UNNEST(@statuses::int[]) AS status
) AS upd
WHERE o.id = upd.id;

-- name: UpdateOrderAccruals :exec
UPDATE order_accruals AS oa
SET accrual = CASE WHEN upd.has_accrual THEN upd.accrual ELSE oa.accrual END
FROM (
    SELECT
        UNNEST(@order_ids::int[]) AS order_id,
        UNNEST(@accruals::real[]) AS accrual,
        UNNEST(@has_accrual::boolean[]) AS has_accrual
) AS upd
WHERE oa.order_id = upd.order_id;
