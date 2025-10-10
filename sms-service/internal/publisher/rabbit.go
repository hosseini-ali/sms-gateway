package publisher

import (
	"context"
	"encoding/json"
	"notif/internal/models"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type RabbitPublisher struct {
	conn *amqp091.Connection
}

func NewRabbitPublisher(conn *amqp091.Connection) *RabbitPublisher {
	return &RabbitPublisher{
		conn: conn,
	}
}

func (r *RabbitPublisher) Publish(ctx context.Context, event models.SMSLog) error {
	channel, err := r.conn.Channel()
	if err != nil {
		return err
	}

	queue := "normal"
	if event.IsExpress {
		queue = "express"
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return channel.PublishWithContext(
		ctx,
		"",    // exchange (empty means default)
		queue, // routing key (queue name)
		false, // mandatory
		false, // immediate
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        []byte(body),
			Timestamp:   time.Now(),
		},
	)
}
