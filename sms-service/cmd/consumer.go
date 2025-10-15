package cmd

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"slices"
	"syscall"
	"time"

	"notif/internal/app"
	"notif/internal/models"
	"notif/internal/repo"

	"notif/config"

	"github.com/spf13/cobra"
)

var consumerName string

func init() {
	consumeCmd.Flags().String("consumer", "", "Name of the RabbitMQ queue to consume from")
	consumeCmd.Flags().Int("workers", 0, "Number of concurrent workers")
}

var consumeCmd = &cobra.Command{
	Use:   "consume",
	Short: "Start consuming messages from a RabbitMQ queue",
	RunE: func(cmd *cobra.Command, args []string) error {
		// load config
		cnf := config.C

		// setup context with cancellation
		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()

		// handle SIGINT / SIGTERM
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		go func() {
			sig := <-sigCh
			log.Printf("[graceful-shutdown] Received signal: %s", sig)
			cancel()
		}()

		consumerName, _ := cmd.Flags().GetString("consumer")
		cmd.MarkFlagRequired("consumer")
		workerCount, _ := cmd.Flags().GetInt("workers")

		if consumerName == "" {
			panic("queue name cannot be empty")
		}
		if workerCount <= 0 {
			panic("set correct worker count")
		}

		app.WithClickHouse(ctx)
		app.WithRabbitMQ(ctx)

		consumers := cnf.RabbitMQ.Queues
		if !slices.Contains(consumers, consumerName) {
			panic("available options: normal, express")
		}

		channel, err := app.A.Rabbit.Channel()
		if err != nil {
			return err
		}
		defer func() {
			_ = channel.Close()
			log.Println("[graceful-shutdown] RabbitMQ channel closed.")
		}()

		msgs, err := channel.Consume(
			consumerName,
			"",    // consumer tag
			false, // autoAck = false (so we control it)
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return err
		}

		log.Printf("[*] Listening for messages on %s...", consumerName)

		// start workers
		for i := range workerCount {
			go func(id int) {
				for {
					select {
					case msg, ok := <-msgs:
						if !ok {
							log.Printf("[worker-%d] channel closed", id)
							return
						}
						log.Printf("[worker-%d] Received message", id)

						err := RabbitHandler(string(msg.Body))
						if err != nil {
							log.Printf("[worker-%d] failed to handle: %v", id, err)
						}

						if err := msg.Ack(false); err != nil {
							log.Printf("[worker-%d] failed to ack: %v", id, err)
						}
					case <-ctx.Done():
						log.Printf("[worker-%d] stopping gracefully...", id)
						return
					}
				}
			}(i + 1)
		}

		// wait until context canceled
		<-ctx.Done()

		log.Println("[graceful-shutdown] Waiting for cleanup...")
		time.Sleep(2 * time.Second) // optional small delay
		log.Println("[graceful-shutdown] Consumer exited cleanly.")
		return nil
	},
}

func RabbitHandler(msg string) error {
	log.Printf("[Consumer - %s] received message: %s", consumerName, msg)

	var SMSLog models.SMSLog
	err := json.Unmarshal([]byte(msg), &SMSLog)
	if err != nil {
		return err
	}

	smsProvider := repo.NewSMSProvider()
	err = smsProvider.SendSMS(SMSLog.PhoneNumber, SMSLog.Message)
	if err != nil {
		return err
	}

	repo := repo.NewSMSStorage(*app.A.Db)
	err = repo.Persist(context.Background(), []models.SMSLog{SMSLog})
	if err != nil {
		return err
	}

	return nil
}
