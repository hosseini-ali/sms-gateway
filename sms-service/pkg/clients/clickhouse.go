package clients

import (
	"context"
	"fmt"
	"log"

	"github.com/ClickHouse/clickhouse-go/v2"
	"notif/config"
)

func NewClickHouse() (clickhouse.Conn, error) {
	ctx := context.Background()
	cnf := config.C

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", cnf.ClickHouse.Host, cnf.ClickHouse.Port)},
		Auth: clickhouse.Auth{
			Database: cnf.ClickHouse.Name,
			Username: cnf.ClickHouse.User,
			Password: cnf.ClickHouse.Password,
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		Debug: true,
	})

	if err != nil {
		log.Fatal("connect error:", err)
		return nil, err
	}

	if err := conn.Ping(ctx); err != nil {
		log.Fatalf("clickhouse ping failed: %v", err)
		return nil, err
	}

	return conn, nil
}
