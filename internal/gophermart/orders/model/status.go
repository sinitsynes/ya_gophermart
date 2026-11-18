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
	UNKNOWN    = "UNKNOWN"
	NEW        = "NEW"
	PROCESSING = "PROCESSING"
	INVALID    = "INVALID"
	PROCESSED  = "PROCESSED"
)

func (s OrderStatus) String() string {
	switch s {
	case OrderStatusUnknown:
		return UNKNOWN
	case OrderStatusNew:
		return NEW
	case OrderStatusProcessing:
		return PROCESSING
	case OrderStatusInvalid:
		return INVALID
	case OrderStatusProcessed:
		return PROCESSED
	default:
		return UNKNOWN
	}
}

func ToInt32(status OrderStatus) int32 {
	switch strings.ToUpper(status.String()) {
	case NEW:
		return int32(OrderStatusNew)
	case PROCESSING:
		return int32(OrderStatusProcessing)
	case INVALID:
		return int32(OrderStatusInvalid)
	case PROCESSED:
		return int32(OrderStatusProcessed)
	default:
		return int32(OrderStatusUnknown)
	}
}

func (s OrderStatus) MarshalJSON() ([]byte, error) {
	return []byte(`"` + s.String() + `"`), nil
}
