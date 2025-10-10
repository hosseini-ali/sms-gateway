package storage

import (
	"context"
	"credit/internal/event"
)

type Storage interface {
	SaveEvents(ctx context.Context, events []event.Event, latestStreamId string, batchSize int) error
	LoadAllBalances(ctx context.Context) (map[string]int, error)
	LatestPersistState(ctx context.Context) (string, error)
}
