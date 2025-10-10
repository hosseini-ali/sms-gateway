package persistance

import (
	"context"
	"credit/internal/event"
	"credit/internal/storage"
)

type Persistor struct {
	store  storage.Storage
	stream event.Stream
}

func NewPersistor(store storage.Storage, stream event.Stream) *Persistor {
	return &Persistor{store, stream}
}

func (p *Persistor) Persist(ctx context.Context, batchSize int64) error {
	latestStreamId, err := p.store.LatestPersistState(ctx)
	if err != nil {
		return err
	}

	events, latestStreamId, err := p.stream.Read(ctx, latestStreamId, batchSize)
	if err != nil {
		return err
	}

	if err = p.store.SaveEvents(ctx, events, latestStreamId, int(batchSize)); err != nil {
		return err
	}

	return nil
}
