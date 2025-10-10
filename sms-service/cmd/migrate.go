package cmd

import (
	"context"
	"fmt"
	"log"

	db "notif/pkg/clients"

	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run ClickHouse migrations",
	Run: func(cmd *cobra.Command, args []string) {
		runMigration()
	},
}

func init() {
	RootCmd.AddCommand(migrateCmd)
}

func runMigration() {
	ctx := context.Background()

	conn, err := db.NewClickHouse()
	if err != nil {
		log.Fatalf("Failed to connect to ClickHouse: %v", err)
	}
	defer conn.Close()

	query := `
	CREATE TABLE IF NOT EXISTS sms_logs (
		phone_number String,
		is_express UInt8,
		org String,
		created_at DateTime DEFAULT now()
	) ENGINE = MergeTree()
	ORDER BY (created_at)
	`

	if err := conn.Exec(ctx, query); err != nil {
		log.Fatalf("Failed to run migration: %v", err)
	}

	fmt.Println("Migration completed: sms_logs table is ready")
}
