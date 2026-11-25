package model

import "strings"

type AccrualStatus int

const (
	AccrualUnknown AccrualStatus = iota
	AccrualRegistered
	AccrualInvalid
	AccrualProcessing
	AccrualProcessed
)

const (
	AccrualStatusUnknown    = "UNKNOWN"
	AccrualStatusRegistered = "REGISTERED"
	AccrualStatusProcessing = "PROCESSING"
	AccrualStatusInvalid    = "INVALID"
	AccrualStatusProcessed  = "PROCESSED"
)

type Accrual struct {
	Order  string   `json:"order"`
	Status string   `json:"status"`
	Points *float32 `json:"accrual,omitempty"`
}

func (a AccrualStatus) String() string {
	switch a {
	case AccrualUnknown:
		return AccrualStatusUnknown
	case AccrualRegistered:
		return AccrualStatusRegistered
	case AccrualProcessing:
		return AccrualStatusProcessing
	case AccrualInvalid:
		return AccrualStatusInvalid
	case AccrualProcessed:
		return AccrualStatusProcessed
	default:
		return AccrualStatusUnknown
	}
}

func AccrualStatusToInt32(status AccrualStatus) int32 {
	switch strings.ToUpper(status.String()) {
	case AccrualStatusRegistered:
		return int32(AccrualRegistered)
	case AccrualStatusProcessing:
		return int32(AccrualProcessing)
	case AccrualStatusInvalid:
		return int32(AccrualInvalid)
	case AccrualStatusProcessed:
		return int32(AccrualProcessed)
	default:
		return int32(AccrualUnknown)
	}
}
