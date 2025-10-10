package persistance_test

import (
	"context"
	"credit/internal/event"
	"credit/internal/persistance"
	"testing"
)

type storage struct {
	events         []event.Event
	latestStreamId string
}

func newStorage() *storage {
	return &storage{
		events: make([]event.Event, 0),
	}
}

func (s *storage) SaveEvents(_ context.Context, events []event.Event, latestStreamId string, batchSize int) error {
	s.events = append(s.events, events...)
	return nil
}

func (s *storage) LoadAllBalances(ctx context.Context) (map[string]int, error) {
	balances := make(map[string]int)
	for _, e := range s.events {
		if e.Type == event.TypeDebit {
			balances[e.OrgID] -= e.Amount
		} else if e.Type == event.TypeCredit {
			balances[e.OrgID] += e.Amount
		}
		s.latestStreamId = e.StreamId
	}
	return balances, nil
}

func (s *storage) LatestPersistState(ctx context.Context) (string, error) {
	return s.latestStreamId, nil
}

type stream struct {
	events []event.Event
}

func newStream(events []event.Event) *stream {
	return &stream{
		events: events,
	}
}

func (s stream) Read(ctx context.Context, from string, count int64) (event []event.Event, latestStreamId string, err error) {
	return s.events, s.events[len(s.events)-1].StreamId, nil
}

func TestPersistor(t *testing.T) {
	persisted := []event.Event{
		{Type: event.TypeCredit, OrgID: "org1", Amount: 1000, StreamId: "foo-1"},
		{Type: event.TypeCredit, OrgID: "org2", Amount: 500, StreamId: "foo-2"},
		{Type: event.TypeDebit, OrgID: "org2", Amount: 200, StreamId: "foo-3"},
		{Type: event.TypeDebit, OrgID: "org1", Amount: 500, StreamId: "foo-4"},
	}
	unseen := []event.Event{
		{Type: event.TypeDebit, OrgID: "org1", Amount: 100, StreamId: "foo-5"},
		{Type: event.TypeDebit, OrgID: "org1", Amount: 100, StreamId: "foo-6"},
		{Type: event.TypeDebit, OrgID: "org2", Amount: 100, StreamId: "foo-7"},
	}

	store := newStorage()
	st := newStream(unseen)

	err := store.SaveEvents(t.Context(), persisted, "", 100)
	if err != nil {
		t.Fatal(err)
	}

	p := persistance.NewPersistor(store, st)
	if err = p.Persist(t.Context(), 100); err != nil {
		t.Fatal(err)
	}

	if len(store.events) != len(persisted)+len(unseen) {
		t.Errorf("expected %d events, got %d", len(persisted)+len(unseen), len(store.events))
	}
}
