package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	repoErr "github.com/sinitsynes/ya_gophermart/internal/err/repository"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/orders/model"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/orders/repository/sqlc"
)

type Storage struct {
	queries *sqlc.Queries
}

func NewStorage(pool *pgxpool.Pool) *Storage {
	return &Storage{
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
			Number:     order.Number,
			Status:     model.OrderStatus(order.Status),
			Accrual:    order.Accrual,
			UploadedAt: order.UploadedAt,
		}
	}
	return results, nil
}

func (s *Storage) CreateOrder(ctx context.Context, order model.CreateOrder) (model.Order, error) {
	created, err := s.queries.CreateOrder(ctx, sqlc.CreateOrderParams{
		UserID: order.UserID,
		Number: order.Number,
		Status: model.OrderStatusToInt32(order.Status),
	})
	if err != nil {
		return model.Order{}, mapError(err)
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
	numbers := make([]string, len(processed))
	statuses := make([]int32, len(processed))
	accruals := make([]float32, len(processed))
	hasAccrual := make([]bool, len(processed))

	for i, order := range processed {
		numbers[i] = order.Number
		statuses[i] = model.OrderStatusToInt32(order.Status)
		if order.Accrual != nil {
			accruals[i] = float32(*order.Accrual)
			hasAccrual[i] = true
		}
	}

	return s.queries.UpdateOrdersAccrual(ctx, sqlc.UpdateOrdersAccrualParams{
		Numbers:    numbers,
		Statuses:   statuses,
		Accruals:   accruals,
		HasAccrual: hasAccrual,
	})
}
