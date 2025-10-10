package cmd

import (
	"credit/internal/conf"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	gomigrate "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
)

var (
	steps           int
	migrationsPath  string
	migrationsTable string
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Long:  "Run database migrations",
	Run:   migrate,
}

func init() {
	migrateCmd.Flags().StringVarP(
		&migrationsPath,
		"migrations-path",
		"m",
		"migrations",
		"path to migrations directory",
	)

	migrateCmd.Flags().StringVarP(
		&migrationsTable,
		"migrations-table",
		"t",
		"migrations",
		"database table holding migrations",
	)

	migrateCmd.Flags().IntVarP(
		&steps,
		"steps",
		"n",
		0,
		"number of steps to migrate. positive steps for up and negative steps for down. zero to upgrade all.",
	)
}

func migrate(_ *cobra.Command, _ []string) {
	if migrationsPath == "" {
		panic("migrations path required")
	}
	if !(strings.HasPrefix(migrationsPath, "/")) {
		path, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		migrationsPath, err = filepath.Abs(filepath.Join(path, migrationsPath))
		if err != nil {
			panic(err)
		}
	}

	db, err := sql.Open("postgres", conf.Cfg.DB.Dsn())
	if err != nil {
		panic(err)
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		panic(err)
	}

	m, err := gomigrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"credit-svc",
		driver,
	)
	if err != nil {
		panic(err)
	}

	if err = m.Up(); err != nil {
		if !errors.Is(err, gomigrate.ErrNoChange) {
			panic(err)
		}
		fmt.Println(err.Error())
	}
}
