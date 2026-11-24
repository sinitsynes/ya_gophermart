package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sinitsynes/ya_gophermart/internal/gophermart/orders/model"
)

const (
	updateBatchSize     = 50
	updateFlushInterval = time.Second
	workers             = 5
	inboxSize           = 256
	defaultRetryAfter   = time.Second
)

var (
	errNoContent = errors.New("no content")
)

type rateLimitError struct {
	retryAfter time.Duration
}

func (e *rateLimitError) Error() string {
	return fmt.Sprintf("accrual rate limited, retry after %s", e.retryAfter)
}

type AccrualRepository interface {
	UpdateOrdersAccrual(context.Context, []model.UpdateOrder) error
}

type AccrualService struct {
	repo        AccrualRepository
	client      *http.Client
	accrualURL  string
	inboxCh     chan string
	outboxCh    chan model.UpdateOrder
	mu          sync.Mutex
	pausedUntil time.Time
}

func NewAccrualService(repo AccrualRepository, url string) *AccrualService {
	return &AccrualService{
		repo: repo,
		client: &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
			},
		},
		accrualURL: strings.TrimRight(url, "/"),
		inboxCh:    make(chan string, inboxSize),
		outboxCh:   make(chan model.UpdateOrder, updateBatchSize),
	}
}

func (a *AccrualService) Start(ctx context.Context) {
	for range workers {
		go a.work(ctx)
	}
	go a.flushLoop(ctx)
}

func (a *AccrualService) EnqueueOrder(orderNumber string) {
	a.inboxCh <- orderNumber
}

func (a *AccrualService) processNumber(ctx context.Context, number string) {
	for {
		if err := a.waitIfRateLimited(ctx); err != nil {
			return
		}

		accrual, err := a.fetchOrderAccrual(ctx, number)
		if err != nil {
			var rl *rateLimitError
			switch {
			case errors.As(err, &rl):
				a.pauseFor(rl.retryAfter)
				continue
			case errors.Is(err, errNoContent):
				if !a.sleep(ctx, defaultRetryAfter) {
					return
				}
				continue
			default:
				if !a.sleep(ctx, defaultRetryAfter) {
					return
				}
				continue
			}
		}

		status := model.ParseOrderStatus(accrual.Status)
		update := model.UpdateOrder{
			Number:  number,
			Status:  status,
			Accrual: accrual.Points,
		}
		select {
		case a.outboxCh <- update:
		case <-ctx.Done():
			return
		}

		if status == model.OrderStatusProcessed || status == model.OrderStatusInvalid {
			return
		}
		if !a.sleep(ctx, defaultRetryAfter) {
			return
		}
	}
}

func (a *AccrualService) sleep(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		return true
	case <-ctx.Done():
		return false
	}
}

func (a *AccrualService) work(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case number := <-a.inboxCh:
			a.processNumber(ctx, number)
		}
	}
}

// flushLoop отправляет пакеты обновлений в репозиторий
func (a *AccrualService) flushLoop(ctx context.Context) {
	buffer := make([]model.UpdateOrder, 0, updateBatchSize)
	ticker := time.NewTicker(updateFlushInterval)
	defer ticker.Stop()

	flush := func() {
		if len(buffer) == 0 {
			return
		}
		_ = a.repo.UpdateOrdersAccrual(ctx, buffer)
		buffer = buffer[:0]
	}

	for {
		select {
		case item, ok := <-a.outboxCh:
			if !ok {
				flush()
				return
			}
			buffer = append(buffer, item)
			if len(buffer) == updateBatchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		case <-ctx.Done():
			flush()
			return
		}
	}
}

// fetchOrderAccrual отправляет запрос к сервису расчета баллов лояльности и возвращает результат
func (a *AccrualService) fetchOrderAccrual(ctx context.Context, orderNumber string) (*model.Accrual, error) {
	url := a.accrualURL + "/api/orders/" + orderNumber
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		accrualResponse := new(model.Accrual)
		if err := json.Unmarshal(data, accrualResponse); err != nil {
			return nil, err
		}
		return accrualResponse, nil
	case http.StatusNoContent:
		return nil, errNoContent
	case http.StatusTooManyRequests:
		return nil, &rateLimitError{retryAfter: parseRetryAfter(resp.Header.Get("Retry-After"))}
	default:
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil, fmt.Errorf("accrual unexpected status %d", resp.StatusCode)
	}
}

// pauseFor устанавливает паузу воркера на указанное время, полученное из заголовка Retry-After
func (a *AccrualService) pauseFor(d time.Duration) {
	until := time.Now().Add(d)
	a.mu.Lock()
	if until.After(a.pausedUntil) {
		a.pausedUntil = until
	}
	a.mu.Unlock()
}

func (a *AccrualService) waitIfRateLimited(ctx context.Context) error {
	a.mu.Lock()
	until := a.pausedUntil
	a.mu.Unlock()

	if !a.sleep(ctx, time.Until(until)) {
		return ctx.Err()
	}
	return nil
}

func parseRetryAfter(v string) time.Duration {
	if v == "" {
		return defaultRetryAfter
	}
	seconds, err := strconv.Atoi(v)
	if err != nil || seconds < 0 {
		return defaultRetryAfter
	}
	return time.Duration(seconds) * time.Second
}
