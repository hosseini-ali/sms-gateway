package storage

import (
	"context"
	"credit/internal/event"
	"database/sql"
	"fmt"
	"strings"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(db *sql.DB) *PostgresStorage {
	return &PostgresStorage{db}
}

func (p *PostgresStorage) LoadAllBalances(ctx context.Context) (map[string]int, error) {
	rows, err := p.db.QueryContext(ctx, "SELECT org_id, amount FROM balances")
	if err != nil {
		return nil, err
	}

	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	balances := make(map[string]int)
	for rows.Next() {
		var orgId string
		var amount int
		if err := rows.Scan(&orgId, &amount); err != nil {
			return nil, err
		}
		balances[orgId] = amount
	}
	return balances, nil
}

func (p *PostgresStorage) SaveLatestPersistState(ctx context.Context, latestStreamId string) error {
	_, err := p.db.ExecContext(ctx, `INSERT INTO persist_state (last_stream_id) VALUES ($1)`, latestStreamId)
	return err
}

func (p *PostgresStorage) SaveBalances(ctx context.Context, balances map[string]int, batchSize int) error {
	if len(balances) == 0 {
		return nil
	}

	orgIDs := make([]string, 0, len(balances))
	for orgID := range balances {
		orgIDs = append(orgIDs, orgID)
	}

	for start := 0; start < len(orgIDs); start += batchSize {
		end := start + batchSize
		if end > len(orgIDs) {
			end = len(orgIDs)
		}

		subset := orgIDs[start:end]
		if err := p.upsertBalancesBatch(ctx, subset, balances); err != nil {
			return fmt.Errorf("upsert batch (%d-%d): %w", start, end, err)
		}
	}

	return nil
}

func (p *PostgresStorage) SaveEvents(ctx context.Context, events []event.Event, latestStreamId string, batchSize int) error {
	if len(events) == 0 {
		return nil
	}

	tx, err := p.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	balances, err := p.LoadAllBalances(ctx)
	if err != nil {
		return err
	}

	for start := 0; start < len(events); start += batchSize {
		end := start + batchSize
		if end > len(events) {
			end = len(events)
		}
		if err := p.insertEventBatch(ctx, events[start:end]); err != nil {
			return fmt.Errorf("insert batch (%d-%d): %w", start, end, err)
		}
	}

	for _, e := range events {
		switch e.Type {
		case event.TypeDebit:
			balances[e.OrgID] -= e.Amount
		case event.TypeCredit:
			balances[e.OrgID] += e.Amount
		default:
			continue
		}
	}

	if err = p.SaveBalances(ctx, balances, batchSize); err != nil {
		return err
	}

	latestStreamId = events[len(events)-1].StreamId

	if err = p.SaveLatestPersistState(ctx, latestStreamId); err != nil {
		return err
	}

	return nil
}

func (p *PostgresStorage) LatestPersistState(ctx context.Context) (string, error) {
	row := p.db.QueryRowContext(ctx, `SELECT last_stream_id FROM persist_state ORDER BY id DESC LIMIT 1`)
	if err := row.Err(); err != nil {
		return "", err
	}
	var lastStream string
	if err := row.Scan(&lastStream); err != nil {
		return "", err
	}
	return lastStream, nil
}

func (p *PostgresStorage) upsertBalancesBatch(ctx context.Context, orgIDs []string, balances map[string]int) error {
	var (
		sb     strings.Builder
		args   []interface{}
		argIdx int
	)

	sb.WriteString(`INSERT INTO balances (org_id, amount) VALUES `)

	for i, orgID := range orgIDs {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fmt.Sprintf("($%d,$%d)", argIdx+1, argIdx+2))
		argIdx += 2
		args = append(args, orgID, balances[orgID])
	}

	// Update existing rows on conflict
	sb.WriteString(` ON CONFLICT (org_id) DO UPDATE SET amount = EXCLUDED.amount`)

	fmt.Println(sb.String(), args)

	_, err := p.db.ExecContext(ctx, sb.String(), args...)
	return err
}

func (p *PostgresStorage) insertEventBatch(ctx context.Context, events []event.Event) error {
	var (
		sb     strings.Builder
		args   []interface{}
		argIdx int
	)

	sb.WriteString(`INSERT INTO credit_events (stream_id, org_id, type, amount) VALUES `)

	for i, e := range events {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(
			fmt.Sprintf("($%d,$%d,$%d,$%d)", argIdx+1, argIdx+2, argIdx+3, argIdx+4),
		)
		argIdx += 4

		args = append(args,
			e.StreamId,
			e.OrgID,
			e.Type,
			e.Amount,
		)
	}

	sb.WriteString(` ON CONFLICT (stream_id) DO NOTHING`)

	_, err := p.db.ExecContext(ctx, sb.String(), args...)
	return err
}
