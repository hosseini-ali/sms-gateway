package cmd

import (
	"context"
	"credit/internal/conf"
	"credit/internal/event"
	"credit/internal/manager"
	"credit/internal/server"
	"credit/internal/storage"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start HTTP server",
	Run:   startFunc,
}

func startFunc(c *cobra.Command, _ []string) {
	rdb := redis.NewClient(&redis.Options{Addr: conf.Cfg.Redis.Dsn()})
	db, err := sql.Open("postgres", conf.Cfg.DB.Dsn())
	if err != nil {
		panic(err)
	}

	eventLogger := event.NewRedisLogger(rdb, "balances:events")
	mgr := manager.NewManager(eventLogger)
	store := storage.NewPostgresStorage(db)
	if err := mgr.LoadInitialBalances(c.Context(), store); err != nil {
		log.Fatalf("failed to load balances: %v", err)
	}

	h := server.NewHandler(mgr)
	srv := server.NewServer(":8081", h)

	go func() {
		log.Println("starting credit service on :8081")
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()
	if err = srv.Shutdown(ctx); err != nil {
		return
	}
}
