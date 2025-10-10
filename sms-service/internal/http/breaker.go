package http

import (
	"sync"
	"time"

	"github.com/sony/gobreaker"
)

var (
	cbMu        sync.Mutex
	orgBreakers = make(map[string]*gobreaker.CircuitBreaker)
)

func getOrgBreaker(org string) *gobreaker.CircuitBreaker {
	cbMu.Lock()
	defer cbMu.Unlock()

	if cb, exists := orgBreakers[org]; exists {
		return cb
	}

	// Create new breaker for this org
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        org,
		MaxRequests: 1, // allow 1 request during half-open(after time out).
		Timeout:     5 * time.Minute,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 3 // open with 3 failure requests
		},
	})
	orgBreakers[org] = cb
	return cb
}
