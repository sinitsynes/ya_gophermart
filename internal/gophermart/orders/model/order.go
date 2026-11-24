package model

import (
	"time"
)

type CreateOrder struct {
	UserID int32
	Number string
	Status OrderStatus
}

type UpdateOrder struct {
	Number  string
	Status  OrderStatus
	Accrual *float32
}

type Order struct {
	ID         int32       `json:"-"`
	UserID     int32       `json:"-"`
	Number     string      `json:"number"`
	Status     OrderStatus `json:"status"`
	Accrual    *float32    `json:"accrual,omitempty"`
	UploadedAt time.Time   `json:"uploaded_at"`
}
