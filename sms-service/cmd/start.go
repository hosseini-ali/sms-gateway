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
	ctx := app.WithGracefulShutdown()

	app.WithClickHouse(ctx)

	app.WithRabbitMQ(ctx)

	app.WithPublisher(ctx)

	app.WithCreditSrv(ctx)

	internal_http.
		NewServer().
		Serve(ctx)

	app.Wait(ctx)
}
