package clients

import (
	"context"
	"log"
	"notif/config"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type Rabbit struct {
	Conn    *amqp091.Connection
	Channel *amqp091.Channel
}

func NewRabbitMQ(url string) (*Rabbit, error) {
	cnf := config.C
	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	// Declare queues
	queues := cnf.RabbitMQ.Queues
	for _, qName := range queues {
		_, err := ch.QueueDeclare(
			qName,
			true,  // durable
			false, // autoDelete
			false, // exclusive
			false, // noWait
			nil,   // args
		)
		if err != nil {
			return nil, err
		}
	}

	return &Rabbit{Conn: conn, Channel: ch}, nil
}

func (r *Rabbit) Close() {
	r.Channel.Close()
	r.Conn.Close()
}

func (r *Rabbit) Publish(ctx context.Context, queue string, body string) error {
	return r.Channel.PublishWithContext(
		ctx,
		"",    // exchange (empty means default)
		queue, // routing key (queue name)
		false, // mandatory
		false, // immediate
		amqp091.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
			Timestamp:   time.Now(),
		},
	)
}

func (r *Rabbit) Consume(queue string, workerCount int, handler func(string)) error {
	msgs, err := r.Channel.Consume(
		queue,
		"",    // consumer name
		true,  // autoAck
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return err
	}

	log.Printf("[*] Listening for messages on %s...", queue)

	for i := range workerCount {

		go func(id int) {

			for msg := range msgs {
				log.Printf("[worker-%d] Received message", id)

				handler(string(msg.Body))
				if err := msg.Ack(false); err != nil {
					log.Printf("[worker-%d] failed to ack message: %v", id, err)
				}
			}
		}(i + 1)
	}

	select {}
}
