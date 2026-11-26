package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sinitsynes/ya_gophermart/internal/gophermart/orders/model"
)

type mockAccrualRepo struct {
	mu      sync.Mutex
	updates [][]model.UpdateOrder
}

func (m *mockAccrualRepo) UpdateOrdersAccrual(_ context.Context, updates []model.UpdateOrder) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := append([]model.UpdateOrder(nil), updates...)
	m.updates = append(m.updates, cp)
	return nil
}

func (m *mockAccrualRepo) allUpdates() []model.UpdateOrder {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []model.UpdateOrder
	for _, batch := range m.updates {
		out = append(out, batch...)
	}
	return out
}

func TestParseRetryAfter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want time.Duration
	}{
		{name: "empty", in: "", want: defaultRetryAfter},
		{name: "seconds", in: "5", want: 5 * time.Second},
		{name: "zero", in: "0", want: 0},
		{name: "invalid", in: "soon", want: defaultRetryAfter},
		{name: "negative", in: "-1", want: defaultRetryAfter},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.want, parseRetryAfter(test.in))
		})
	}
}

func TestFetchOrderAccrual(t *testing.T) {
	t.Parallel()

	points := float32(500)

	tests := []struct {
		name          string
		handler       http.HandlerFunc
		want          *model.Accrual
		wantErr       error
		wantRateLimit bool
		wantStatus    int
	}{
		{
			name: "OK",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(model.Accrual{
					Order:  "123",
					Status: model.OrderStatusProcessed,
					Points: &points,
				})
			},
			want: &model.Accrual{
				Order:  "123",
				Status: model.OrderStatusProcessed,
				Points: &points,
			},
		},
		{
			name: "NoContent",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			},
			wantErr: errNoContent,
		},
		{
			name: "TooManyRequests",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Retry-After", "3")
				w.WriteHeader(http.StatusTooManyRequests)
			},
			wantRateLimit: true,
		},
		{
			name: "UnexpectedStatus",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(test.handler)
			t.Cleanup(srv.Close)

			a := NewAccrualService(nil, srv.URL)
			got, err := a.fetchOrderAccrual(t.Context(), "123")

			switch {
			case test.wantErr != nil:
				assert.ErrorIs(t, err, test.wantErr)
			case test.wantRateLimit:
				var rl *rateLimitError
				assert.ErrorAs(t, err, &rl)
				assert.Equal(t, 3*time.Second, rl.retryAfter)
			case test.wantStatus != 0:
				assert.Error(t, err)
				assert.ErrorContains(t, err, "unexpected status")
			default:
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, test.want.Order, got.Order)
				assert.Equal(t, test.want.Status, got.Status)
				assert.NotNil(t, got.Points)
				assert.Equal(t, *test.want.Points, *got.Points)
			}
		})
	}
}

func TestAccrualProcessesProcessedOrder(t *testing.T) {
	t.Parallel()
	orderNumber := "4561261212345631"

	points := float32(42.5)
	repo := &mockAccrualRepo{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/orders/4561261212345631", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(model.Accrual{
			Order:  orderNumber,
			Status: model.OrderStatusProcessed,
			Points: &points,
		})
	}))
	t.Cleanup(srv.Close)

	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)

	a := NewAccrualService(repo, srv.URL)
	a.Start(ctx)
	a.EnqueueOrder(7, orderNumber)

	require.Eventually(t, func() bool {
		return len(repo.allUpdates()) > 0
	}, 3*time.Second, 50*time.Millisecond)

	updates := repo.allUpdates()
	require.Len(t, updates, 1)
	assert.Equal(t, int32(7), updates[0].ID)
	assert.Equal(t, orderNumber, updates[0].Number)
	assert.Equal(t, model.OrderProcessed, updates[0].Status)
	require.NotNil(t, updates[0].Accrual)
	assert.Equal(t, points, *updates[0].Accrual)
}

func TestAccrualRetriesAfterRateLimit(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	repo := &mockAccrualRepo{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := calls.Add(1)
		if n == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(model.Accrual{
			Order:  "1",
			Status: model.OrderStatusInvalid,
		})
	}))
	t.Cleanup(srv.Close)

	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)

	a := NewAccrualService(repo, srv.URL)
	a.Start(ctx)
	a.EnqueueOrder(1, "1")

	require.Eventually(t, func() bool {
		return len(repo.allUpdates()) > 0
	}, 3*time.Second, 50*time.Millisecond)

	assert.GreaterOrEqual(t, calls.Load(), int32(2))
	updates := repo.allUpdates()
	require.NotEmpty(t, updates)
	assert.Equal(t, model.OrderInvalid, updates[0].Status)
	assert.Nil(t, updates[0].Accrual)
}

func TestPauseForAndWaitIfRateLimited(t *testing.T) {
	t.Parallel()

	a := NewAccrualService(nil, "http://localhost")
	a.pauseFor(30 * time.Millisecond)

	start := time.Now()
	err := a.waitIfRateLimited(t.Context())
	require.NoError(t, err)
	assert.GreaterOrEqual(t, time.Since(start), 20*time.Millisecond)
}
