package cmd

import (
	"credit/internal/conf"
	"credit/internal/event"
	"credit/internal/persistance"
	"credit/internal/storage"
	"database/sql"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
)

var (
	batch    int
	stream   string
	group    string
	consumer string
)

var persistCmd = &cobra.Command{
	Use:   "persist",
	Short: "persist events to the database",
	Run:   persist,
}

func init() {
	persistCmd.Flags().IntVarP(
		&batch,
		"batch",
		"n",
		1000,
		"number of events to persist in oen part",
	)

	persistCmd.Flags().StringVarP(
		&stream,
		"stream",
		"s",
		"foo",
		"name of the stream to read from",
	)

	persistCmd.Flags().StringVarP(
		&group,
		"group",
		"g",
		"gp",
		"name of the group",
	)

	persistCmd.Flags().StringVarP(
		&consumer,
		"consumer",
		"c",
		"c",
		"name of the consumer",
	)
}

func persist(cmd *cobra.Command, _ []string) {
	if batch == 0 {
		return
	}
	db, err := sql.Open("postgres", conf.Cfg.DB.Dsn())
	if err != nil {
		panic(err)
	}

	rdb := redis.NewClient(&redis.Options{Addr: conf.Cfg.Redis.Dsn()})

	persistor := persistance.NewPersistor(
		storage.NewPostgresStorage(db),
		event.NewStream(rdb, stream, group, consumer),
	)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	//ctx, cancel := context.WithCancel(cmd.Context())
	//defer cancel()

	if err = persistor.Persist(cmd.Context(), int64(batch)); err != nil {
		panic(err)
	}

}
