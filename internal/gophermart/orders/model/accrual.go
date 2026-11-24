package model

import "strings"

type AccrualStatus int

const (
	AccrualStatusUnknown AccrualStatus = iota
	AccrualStatusRegistered
	AccrualStatusInvalid
	AccrualStatusProcessing
	AccrualStatusProcessed
)

const (
	ACCRUAL_UNKNOWN    = "UNKNOWN"
	ACCRUAL_REGISTERED = "REGISTERED"
	ACCRUAL_PROCESSING = "PROCESSING"
	ACCRUAL_INVALID    = "INVALID"
	ACCRUAL_PROCESSED  = "PROCESSED"
)

type Accrual struct {
	Order  string   `json:"order"`
	Status string   `json:"status"`
	Points *float32 `json:"accrual,omitempty"`
}

func (a AccrualStatus) String() string {
	switch a {
	case AccrualStatusUnknown:
		return ACCRUAL_UNKNOWN
	case AccrualStatusRegistered:
		return ACCRUAL_REGISTERED
	case AccrualStatusProcessing:
		return ACCRUAL_PROCESSING
	case AccrualStatusInvalid:
		return ACCRUAL_INVALID
	case AccrualStatusProcessed:
		return ACCRUAL_PROCESSED
	default:
		return ACCRUAL_UNKNOWN
	}
}

func AccrualStatusToInt32(status AccrualStatus) int32 {
	switch strings.ToUpper(status.String()) {
	case ACCRUAL_REGISTERED:
		return int32(AccrualStatusRegistered)
	case ACCRUAL_PROCESSING:
		return int32(AccrualStatusProcessing)
	case ACCRUAL_INVALID:
		return int32(AccrualStatusInvalid)
	case ACCRUAL_PROCESSED:
		return int32(AccrualStatusProcessed)
	default:
		return int32(OrderStatusUnknown)
	}
}
