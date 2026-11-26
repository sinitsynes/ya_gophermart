package model

import "strings"

type OrderStatus int

const (
	OrderUnknown OrderStatus = iota
	OrderNew
	OrderProcessing
	OrderInvalid
	OrderProcessed
)

const (
	OrderStatusUnknown    = "UNKNOWN"
	OrderStatusNew        = "NEW"
	OrderStatusProcessing = "PROCESSING"
	OrderStatusInvalid    = "INVALID"
	OrderStatusProcessed  = "PROCESSED"
)

func (s OrderStatus) String() string {
	switch s {
	case OrderUnknown:
		return OrderStatusUnknown
	case OrderNew:
		return OrderStatusNew
	case OrderProcessing:
		return OrderStatusProcessing
	case OrderInvalid:
		return OrderStatusInvalid
	case OrderProcessed:
		return OrderStatusProcessed
	default:
		return OrderStatusUnknown
	}
}

func OrderStatusToInt32(status OrderStatus) int32 {
	switch strings.ToUpper(status.String()) {
	case OrderStatusNew:
		return int32(OrderNew)
	case OrderStatusProcessing:
		return int32(OrderProcessing)
	case OrderStatusInvalid:
		return int32(OrderInvalid)
	case OrderStatusProcessed:
		return int32(OrderProcessed)
	default:
		return int32(OrderUnknown)
	}
}

func ParseOrderStatus(status string) OrderStatus {
	switch strings.ToUpper(status) {
	case OrderStatusNew:
		return OrderNew
	case OrderStatusProcessing:
		return OrderProcessing
	case OrderStatusInvalid:
		return OrderInvalid
	case OrderStatusProcessed:
		return OrderProcessed
	default:
		return OrderUnknown
	}
}

func (s OrderStatus) MarshalJSON() ([]byte, error) {
	return []byte(`"` + s.String() + `"`), nil
}
