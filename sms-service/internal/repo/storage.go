package repo

import (
	"context"
	"fmt"
	"notif/internal/models"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
)

type SMSStorage interface {
	Persist(ctx context.Context, s []models.SMSLog) error
	CountByOrg(ctx context.Context, org string) (uint64, error)
}

type ClickHouseStorage struct {
	conn clickhouse.Conn
}

func (c ClickHouseStorage) Persist(ctx context.Context, logs []models.SMSLog) error {

	batch, err := c.conn.PrepareBatch(ctx, "INSERT INTO sms_logs (phone_number, org, is_express, created_at)")
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	for _, sms := range logs {

		var express uint8
		if sms.IsExpress {
			express = 1
		}
		if err := batch.Append(sms.PhoneNumber, sms.Org, express, time.Now()); err != nil {
			return fmt.Errorf("failed to append row: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("failed to send batch: %w", err)
	}

	return nil
}

// CountByOrg returns the number of records for a given organization in sms_logs table.
func (c ClickHouseStorage) CountByOrg(ctx context.Context, org string) (uint64, error) {
	var count uint64
	// Use a parameterized query to count rows for the specific org
	if err := c.conn.QueryRow(ctx, "SELECT count() FROM sms_logs WHERE org = ?", org).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func NewSMSStorage(conn clickhouse.Conn) ClickHouseStorage {
	return ClickHouseStorage{
		conn: conn,
	}
}
