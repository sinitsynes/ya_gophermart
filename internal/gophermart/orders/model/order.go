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
	ID      int32
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

type Balance struct {
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

type CreateWithdrawal struct {
	UserID int32
	Number string
	Sum    float32
}

type Withdrawal struct {
	Order       string    `json:"order"`
	Sum         float32   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}
