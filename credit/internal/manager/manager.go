package manager

import (
	"context"
	"credit/internal/event"
	"credit/internal/storage"
	"errors"
	"sync"
	"time"
)

var (
	ErrInsufficient    = errors.New("insufficient funds")
	ErrAccountNotFound = errors.New("account not found")
)

// Manager manages sharded balances
type Manager struct {
	accounts    map[string]*account
	eventLogger event.Logger
}

type account struct {
	amount int
	mu     sync.Mutex
}

func NewManager(el event.Logger) *Manager {
	m := &Manager{eventLogger: el}
	m.accounts = make(map[string]*account)
	return m
}

func (m *Manager) LoadInitialBalances(ctx context.Context, store storage.Storage) error {
	balances, err := store.LoadAllBalances(ctx)
	if err != nil {
		return err
	}

	for orgID, amt := range balances {
		acc := m.accounts[orgID]
		if acc == nil {
			acc = &account{}
		}
		acc.amount = amt
	}

	return nil
}

// Debit atomically deducts amount; returns remaining balance
func (m *Manager) Debit(ctx context.Context, org string, amount int, txnID string) (int, error) {
	acc := m.accounts[org]
	if acc == nil {
		return 0, ErrAccountNotFound
	}
	acc.mu.Lock()
	if acc.amount < amount {
		acc.mu.Unlock()
		return 0, ErrInsufficient
	}
	acc.amount -= amount
	r := acc.amount
	acc.mu.Unlock()

	// todo: revert amount back if logging encountered error
	_ = m.eventLogger.Log(ctx, event.Event{Type: event.TypeDebit, OrgID: org, Amount: amount, TxnID: txnID, Timestamp: time.Now()})
	return r, nil
}

// Credit adds funds; returns new balance
func (m *Manager) Credit(ctx context.Context, org string, amount int, txnID string) (int, error) {
	acc := m.accounts[org]
	if acc == nil {
		acc = &account{}
	}
	acc.mu.Lock()
	acc.amount += amount
	m.accounts[org] = acc
	acc.mu.Unlock()

	// todo: revert amount back if logging encountered error
	_ = m.eventLogger.Log(ctx, event.Event{Type: event.TypeCredit, OrgID: org, Amount: amount, TxnID: txnID, Timestamp: time.Now()})
	return acc.amount, nil
}

// Balance returns current in-memory balance
func (m *Manager) Balance(_ context.Context, org string) int {
	acc := m.accounts[org]
	if acc == nil {
		return 0
	}
	acc.mu.Lock()
	defer acc.mu.Unlock()

	return acc.amount
}
