package model

import "strings"

type OrderStatus int

const (
	OrderStatusUnknown OrderStatus = iota
	OrderStatusNew
	OrderStatusProcessing
	OrderStatusInvalid
	OrderStatusProcessed
)

const (
	ORDER_UNKNOWN    = "UNKNOWN"
	ORDER_NEW        = "NEW"
	ORDER_PROCESSING = "PROCESSING"
	ORDER_INVALID    = "INVALID"
	ORDER_PROCESSED  = "PROCESSED"
)

func (s OrderStatus) String() string {
	switch s {
	case OrderStatusUnknown:
		return ORDER_UNKNOWN
	case OrderStatusNew:
		return ORDER_NEW
	case OrderStatusProcessing:
		return ORDER_PROCESSING
	case OrderStatusInvalid:
		return ORDER_INVALID
	case OrderStatusProcessed:
		return ORDER_PROCESSED
	default:
		return ORDER_UNKNOWN
	}
}

func OrderStatusToInt32(status OrderStatus) int32 {
	switch strings.ToUpper(status.String()) {
	case ORDER_NEW:
		return int32(OrderStatusNew)
	case ORDER_PROCESSING:
		return int32(OrderStatusProcessing)
	case ORDER_INVALID:
		return int32(OrderStatusInvalid)
	case ORDER_PROCESSED:
		return int32(OrderStatusProcessed)
	default:
		return int32(OrderStatusUnknown)
	}
}

func ParseOrderStatus(status string) OrderStatus {
	switch strings.ToUpper(status) {
	case ORDER_NEW, "REGISTERED":
		return OrderStatusNew
	case ORDER_PROCESSING:
		return OrderStatusProcessing
	case ORDER_INVALID:
		return OrderStatusInvalid
	case ORDER_PROCESSED:
		return OrderStatusProcessed
	default:
		return OrderStatusUnknown
	}
}

func (s OrderStatus) MarshalJSON() ([]byte, error) {
	return []byte(`"` + s.String() + `"`), nil
}
