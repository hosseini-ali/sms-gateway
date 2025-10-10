package event

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type Stream interface {
	Read(ctx context.Context, from string, count int64) (event []Event, latestStreamId string, err error)
}

type RedisStream struct {
	rdb      *redis.Client
	stream   string
	group    string
	consumer string
}

func NewStream(rdb *redis.Client, stream, group, consumer string) *RedisStream {
	return &RedisStream{rdb: rdb, stream: stream, group: group, consumer: consumer}
}

func (s *RedisStream) Read(ctx context.Context, from string, count int64) (events []Event, latestStreamId string, err error) {
	args := &redis.XReadGroupArgs{
		Group:    s.group,
		Consumer: s.consumer,
		Streams:  []string{s.stream, from},
		Count:    count,
		Block:    0,
	}

	streams, err := s.rdb.XReadGroup(ctx, args).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, latestStreamId, nil
		}
		return nil, latestStreamId, fmt.Errorf("xreadgroup: %w", err)
	}

	for _, s := range streams {
		for _, m := range s.Messages {
			latestStreamId = m.ID
			raw, ok := m.Values["event"]
			if !ok {
				continue
			}

			var e Event
			switch v := raw.(type) {
			case string:
				if err := json.Unmarshal([]byte(v), &e); err != nil {
					return events, latestStreamId, fmt.Errorf("decode event (%s): %w", m.ID, err)
				}
			case []byte:
				if err := json.Unmarshal(v, &e); err != nil {
					return events, latestStreamId, fmt.Errorf("decode event (%s): %w", m.ID, err)
				}
			default:
				return events, latestStreamId, fmt.Errorf("unexpected event type %T for id=%s", raw, m.ID)
			}
			e.StreamId = m.ID
			events = append(events, e)
		}
	}

	return events, latestStreamId, nil
}
