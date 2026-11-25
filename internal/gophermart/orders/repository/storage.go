package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	repoErr "github.com/sinitsynes/ya_gophermart/internal/err/repository"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/orders/model"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/orders/repository/sqlc"
)

type Storage struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func NewStorage(pool *pgxpool.Pool) *Storage {
	return &Storage{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

func (s *Storage) GetOrdersByUserID(ctx context.Context, userID int32) ([]model.Order, error) {
	orders, err := s.queries.GetOrdersByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(orders) == 0 {
		return nil, repoErr.ErrNotFound
	}
	results := make([]model.Order, len(orders))
	for i, order := range orders {
		results[i] = model.Order{
			ID:         order.ID,
			UserID:     order.UserID,
			Number:     order.Number,
			Status:     model.OrderStatus(order.Status),
			Accrual:    order.Accrual,
			UploadedAt: order.UploadedAt,
		}
	}
	return results, nil
}

func (s *Storage) CreateOrder(ctx context.Context, order model.CreateOrder) (model.Order, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return model.Order{}, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	qtx := s.queries.WithTx(tx)
	created, err := qtx.CreateOrder(ctx, sqlc.CreateOrderParams{
		UserID: order.UserID,
		Number: order.Number,
		Status: model.OrderStatusToInt32(order.Status),
	})
	if err != nil {
		return model.Order{}, mapError(err)
	}
	if createErr := qtx.CreateOrderAccrual(ctx, created.ID); createErr != nil {
		return model.Order{}, mapError(createErr)
	}
	if txErr := tx.Commit(ctx); txErr != nil {
		return model.Order{}, txErr
	}

	return model.Order{
		ID:         created.ID,
		UserID:     created.UserID,
		Number:     created.Number,
		Status:     model.OrderStatus(created.Status),
		UploadedAt: created.UploadedAt,
	}, nil
}

func (s *Storage) GetOrderByNumber(ctx context.Context, orderNumber string) (model.Order, error) {
	order, err := s.queries.GetOrderByNumber(ctx, orderNumber)
	if err != nil {
		return model.Order{}, mapError(err)
	}
	return model.Order{
		ID:         order.ID,
		UserID:     order.UserID,
		Number:     order.Number,
		Status:     model.OrderStatus(order.Status),
		Accrual:    order.Accrual,
		UploadedAt: order.UploadedAt,
	}, nil
}

func (s *Storage) UpdateOrdersAccrual(ctx context.Context, processed []model.UpdateOrder) error {
	if len(processed) == 0 {
		return nil
	}

	ids := make([]int32, len(processed))
	statuses := make([]int32, len(processed))
	accruals := make([]float32, len(processed))
	hasAccrual := make([]bool, len(processed))

	for i, order := range processed {
		ids[i] = order.ID
		statuses[i] = model.OrderStatusToInt32(order.Status)
		if order.Accrual != nil {
			accruals[i] = *order.Accrual
			hasAccrual[i] = true
		}
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	qtx := s.queries.WithTx(tx)
	if statusesErr := qtx.UpdateOrderStatuses(ctx, sqlc.UpdateOrderStatusesParams{
		Ids:      ids,
		Statuses: statuses,
	}); statusesErr != nil {
		return statusesErr
	}
	if accrualsErr := qtx.UpdateOrderAccruals(ctx, sqlc.UpdateOrderAccrualsParams{
		OrderIds:   ids,
		Accruals:   accruals,
		HasAccrual: hasAccrual,
	}); accrualsErr != nil {
		return accrualsErr
	}
	return tx.Commit(ctx)
}
