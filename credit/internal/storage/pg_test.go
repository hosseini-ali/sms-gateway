package storage_test

import (
	"credit/internal/event"
	"credit/internal/storage"
	"database/sql"
	"errors"
	"reflect"
	"testing"
	"time"

	gomigrate "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func openDb() (*sql.DB, error) {
	return sql.Open("postgres", "postgres://postgres:postgres@0.0.0.0:5432/credit-svc?sslmode=disable")
}

func setup(t *testing.T, db *sql.DB) func() {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		t.Fatal(err)
	}
	m, err := gomigrate.NewWithDatabaseInstance(
		"file://"+"../../migrations",
		"credit-svc",
		driver,
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := m.Up(); err != nil {
		if !errors.Is(err, gomigrate.ErrNoChange) {
			t.Fatal(err)
		}
	}

	return func() {
		if err := m.Down(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestPostgresStorage_LatestPersistState(t *testing.T) {
	db, err := openDb()
	if err != nil {
		t.Fatal(err)
	}
	teardown := setup(t, db)
	defer teardown()

	repo := storage.NewPostgresStorage(db)

	streamId := "foo-123"

	if err := repo.SaveLatestPersistState(t.Context(), streamId); err != nil {
		t.Error(err)
	}

	foundLatestStreamId, err := repo.LatestPersistState(t.Context())
	if err != nil {
		t.Error(err)
	}
	if foundLatestStreamId != streamId {
		t.Errorf("latest persist state was %#v, expected %#v", foundLatestStreamId, streamId)
	}
}

func TestPostgresStorage_SaveBalances(t *testing.T) {
	db, err := openDb()
	if err != nil {
		t.Fatal(err)
	}
	teardown := setup(t, db)
	defer teardown()

	repo := storage.NewPostgresStorage(db)

	balances := map[string]int{
		"org1": 100,
		"org2": 200,
		"org3": 300,
	}

	if err = repo.SaveBalances(t.Context(), balances, 100); err != nil {
		t.Error(err)
	}

	fetchedBalanced, err := repo.LoadAllBalances(t.Context())
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(fetchedBalanced, balances) {
		t.Errorf("expected %#v, got %#v", balances, fetchedBalanced)
	}
}

func TestPostgresStorage_SaveEvents(t *testing.T) {
	db, err := openDb()
	if err != nil {
		t.Fatal(err)
	}
	teardown := setup(t, db)
	defer teardown()
	repo := storage.NewPostgresStorage(db)
	events := []event.Event{
		{Type: event.TypeCredit, OrgID: "org1", Amount: 1000, TxnID: "foo", StreamId: "foo-1", Timestamp: time.Now()},
		{Type: event.TypeDebit, OrgID: "org1", Amount: 100, TxnID: "foo", StreamId: "foo-2", Timestamp: time.Now()},
		{Type: event.TypeDebit, OrgID: "org1", Amount: 200, TxnID: "foo", StreamId: "foo-3", Timestamp: time.Now()},
		{Type: event.TypeDebit, OrgID: "org1", Amount: 300, TxnID: "foo", StreamId: "foo-4", Timestamp: time.Now()},
	}

	if err = repo.SaveEvents(t.Context(), events, "", 100); err != nil {
		t.Error(err)
	}

	balances, err := repo.LoadAllBalances(t.Context())
	if err != nil {
		t.Error(err)
	}
	if balances["org1"] != 400 {
		t.Errorf("expected 400, got %d", balances["org1"])
	}

	latestStreamId, err := repo.LatestPersistState(t.Context())
	if err != nil {
		t.Error(err)
	}
	if latestStreamId != "foo-4" {
		t.Errorf("expected foo-4, got %s", latestStreamId)
	}
}
