package cmd

import (
	"notif/internal/app"
	internal_http "notif/internal/http"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "A sample application that can be used",
	Long:  `A simple application that can be used as a sample that has redis, mariadb and tracing and logging`,
	Run:   startFunc,
}

func startFunc(cmd *cobra.Command, _ []string) {
	app.WithGracefulShutdown()

	app.WithClickHouse(cmd.Context())

	app.WithRabbitMQ(cmd.Context())

	app.WithPublisher(cmd.Context())

	app.WithCreditSrv(cmd.Context())

	internal_http.
		NewServer().
		Serve()

	app.Wait()
}
