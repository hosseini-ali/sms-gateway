package event

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	TypeDebit  = "debit"
	TypeCredit = "credit"
)

// Event represents a balance event
type Event struct {
	Type      string    `json:"type"`
	OrgID     string    `json:"org_id"`
	Amount    int       `json:"amount"`
	TxnID     string    `json:"txn_id"`
	StreamId  string    `json:"stream_id"`
	Timestamp time.Time `json:"ts"`
}

// Logger is an interface for event sinks
type Logger interface {
	Log(ctx context.Context, e Event) error
}

// RedisLogger writes events to a Redis RedisStream
type RedisLogger struct {
	rdb    *redis.Client
	stream string
}

func NewRedisLogger(rdb *redis.Client, stream string) *RedisLogger {
	return &RedisLogger{rdb: rdb, stream: stream}
}

func (l *RedisLogger) Log(ctx context.Context, e Event) error {
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}
	_, err = l.rdb.XAdd(ctx, &redis.XAddArgs{Stream: l.stream, Values: map[string]interface{}{"event": string(b)}}).Result()
	return err
}
