package model_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sinitsynes/ya_gophermart/internal/gophermart/orders/model"
)

func TestOrderStatusString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status model.OrderStatus
		want   string
	}{
		{status: model.OrderUnknown, want: model.OrderStatusUnknown},
		{status: model.OrderNew, want: model.OrderStatusNew},
		{status: model.OrderProcessing, want: model.OrderStatusProcessing},
		{status: model.OrderInvalid, want: model.OrderStatusInvalid},
		{status: model.OrderProcessed, want: model.OrderStatusProcessed},
		{status: model.OrderStatus(99), want: model.OrderStatusUnknown},
	}

	for _, test := range tests {
		t.Run(test.want, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.want, test.status.String())
		})
	}
}

func TestOrderStatusToInt32(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status model.OrderStatus
		want   int32
	}{
		{name: "New", status: model.OrderNew, want: int32(model.OrderNew)},
		{name: "Processing", status: model.OrderProcessing, want: int32(model.OrderProcessing)},
		{name: "Invalid", status: model.OrderInvalid, want: int32(model.OrderInvalid)},
		{name: "Processed", status: model.OrderProcessed, want: int32(model.OrderProcessed)},
		{name: "Unknown", status: model.OrderUnknown, want: int32(model.OrderUnknown)},
		{name: "OutOfRange", status: model.OrderStatus(99), want: int32(model.OrderUnknown)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.want, model.OrderStatusToInt32(test.status))
		})
	}
}

func TestParseOrderStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  model.OrderStatus
	}{
		{name: "New", input: "NEW", want: model.OrderNew},
		{name: "NewLower", input: "new", want: model.OrderNew},
		{name: "Processing", input: "PROCESSING", want: model.OrderProcessing},
		{name: "Invalid", input: "INVALID", want: model.OrderInvalid},
		{name: "Processed", input: "PROCESSED", want: model.OrderProcessed},
		{name: "Unknown", input: "UNKNOWN", want: model.OrderUnknown},
		{name: "Empty", input: "", want: model.OrderUnknown},
		{name: "Garbage", input: "something", want: model.OrderUnknown},
		{name: "Registered", input: "REGISTERED", want: model.OrderUnknown},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.want, model.ParseOrderStatus(test.input))
		})
	}
}

func TestOrderStatusMarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status model.OrderStatus
		want   string
	}{
		{name: "New", status: model.OrderNew, want: `"NEW"`},
		{name: "Processing", status: model.OrderProcessing, want: `"PROCESSING"`},
		{name: "Invalid", status: model.OrderInvalid, want: `"INVALID"`},
		{name: "Processed", status: model.OrderProcessed, want: `"PROCESSED"`},
		{name: "Unknown", status: model.OrderUnknown, want: `"UNKNOWN"`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := json.Marshal(test.status)
			require.NoError(t, err)
			assert.Equal(t, test.want, string(got))
		})
	}
}
