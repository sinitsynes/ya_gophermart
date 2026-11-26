package model_test

import (
	"encoding/json/v2"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sinitsynes/ya_gophermart/internal/gophermart/orders/model"
)

func TestAccrualStatusString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status model.AccrualStatus
		want   string
	}{
		{status: model.AccrualUnknown, want: model.AccrualStatusUnknown},
		{status: model.AccrualRegistered, want: model.AccrualStatusRegistered},
		{status: model.AccrualProcessing, want: model.AccrualStatusProcessing},
		{status: model.AccrualInvalid, want: model.AccrualStatusInvalid},
		{status: model.AccrualProcessed, want: model.AccrualStatusProcessed},
		{status: model.AccrualStatus(99), want: model.AccrualStatusUnknown},
	}

	for _, test := range tests {
		t.Run(test.want, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.want, test.status.String())
		})
	}
}

func TestAccrualStatusToInt32(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status model.AccrualStatus
		want   int32
	}{
		{name: "Registered", status: model.AccrualRegistered, want: int32(model.AccrualRegistered)},
		{name: "Processing", status: model.AccrualProcessing, want: int32(model.AccrualProcessing)},
		{name: "Invalid", status: model.AccrualInvalid, want: int32(model.AccrualInvalid)},
		{name: "Processed", status: model.AccrualProcessed, want: int32(model.AccrualProcessed)},
		{name: "Unknown", status: model.AccrualUnknown, want: int32(model.AccrualUnknown)},
		{name: "OutOfRange", status: model.AccrualStatus(99), want: int32(model.AccrualUnknown)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.want, model.AccrualStatusToInt32(test.status))
		})
	}
}

func TestAccrualJSON(t *testing.T) {
	t.Parallel()

	points := float32(500.5)

	tests := []struct {
		name string
		in   string
		want model.Accrual
	}{
		{
			name: "WithPoints",
			in:   `{"order":"123","status":"PROCESSED","accrual":500.5}`,
			want: model.Accrual{
				Order:  "123",
				Status: model.AccrualStatusProcessed,
				Points: &points,
			},
		},
		{
			name: "WithoutPoints",
			in:   `{"order":"123","status":"PROCESSING"}`,
			want: model.Accrual{
				Order:  "123",
				Status: model.AccrualStatusProcessing,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var got model.Accrual
			require.NoError(t, json.Unmarshal([]byte(test.in), &got))
			assert.Equal(t, test.want.Order, got.Order)
			assert.Equal(t, test.want.Status, got.Status)
			if test.want.Points == nil {
				assert.Nil(t, got.Points)
				return
			}
			require.NotNil(t, got.Points)
			assert.Equal(t, *test.want.Points, *got.Points)
		})
	}
}
