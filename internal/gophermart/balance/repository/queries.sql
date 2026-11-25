-- name: GetBalanceByUser :one
SELECT
    (
        COALESCE((
            SELECT SUM(oa.accrual)
            FROM order_accruals oa
            JOIN orders o ON o.id = oa.order_id
            WHERE o.user_id = @user_id
        ), 0)
        -
        COALESCE((
            SELECT SUM(w.sum)
            FROM order_withdrawals w
            WHERE w.user_id = @user_id
        ), 0)
    )::real AS current,
    COALESCE((
        SELECT SUM(w.sum)
        FROM order_withdrawals w
        WHERE w.user_id = @user_id
    ), 0)::real AS withdrawn;

-- name: CreateWithdrawal :exec
INSERT INTO order_withdrawals (user_id, number, sum)
VALUES (@user_id, @number, @sum);


-- name: GetWithdrawalsByUser :many
SELECT id, user_id, number, sum, processed_at
FROM order_withdrawals
WHERE user_id = @user_id
ORDER BY processed_at DESC;
