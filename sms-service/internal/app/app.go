package app

import (
	"context"
	"fmt"
	"log"
	"notif/config"
	"notif/internal/credit"
	"notif/internal/publisher"
	"os"
	"os/signal"
	"syscall"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/rabbitmq/amqp091-go"
)

type application struct {
	cancelFunc context.CancelFunc
	Db         *clickhouse.Conn
	Rabbit     *amqp091.Connection
	CreditSrv  credit.CreditSrv
	Publisher  publisher.Publisher
}

var (
	// A is the singleton instance of application
	A *application
)

func init() {
	A = &application{}
}

func WithPublisher(ctx context.Context) {
	A.Publisher = publisher.NewRabbitPublisher(A.Rabbit)
}

func WithClickHouse(ctx context.Context) {
	cnf := config.C

	fmt.Println(fmt.Sprint("%+v", cnf))

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
		panic(err)
	}

	if err := conn.Ping(ctx); err != nil {
		log.Fatalf("clickhouse ping failed: %v", err)
		panic(err)
	}

	A.Db = &conn
}

func WithRabbitMQ(ctx context.Context) {
	cnf := config.C
	conn, err := amqp091.Dial(cnf.RabbitMQ.URL)
	if err != nil {
		panic(err)
	}

	ch, err := conn.Channel()
	if err != nil {
		panic(err)
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
			panic(err)
		}
	}

	A.Rabbit = conn
}

func WithCreditSrv(ctx context.Context) {
	A.CreditSrv = credit.NewHttpCreditClient(config.C.CreditSrv.BaseUrl)
}

// WithGracefulShutdown registers a signal handler for graceful shutdown
func WithGracefulShutdown() context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		sig := <-c
		log.Println("system call:", sig)
		cancel()
	}()

	return ctx
}

func Wait(ctx context.Context) {
	<-ctx.Done()
}