package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
	"sync"

	"github.com/Jackob2004/dvz-web/internal/database"
	"github.com/Jackob2004/dvz-web/internal/version"

	"github.com/lmittmann/tint"
)

func main() {
	logger := slog.New(tint.NewTextHandler(os.Stdout, &tint.Options{Level: slog.LevelDebug}))

	err := run(logger)
	if err != nil {
		trace := string(debug.Stack())
		logger.Error(err.Error(), "trace", trace)
		os.Exit(1)
	}
}

type config struct {
	baseURL          string
	httpPort         int
	internalHttpPort int
	db               struct {
		dsn         string
		automigrate bool
	}
}

type application struct {
	config config
	db     *database.DB
	logger *slog.Logger
	wg     sync.WaitGroup
}

func run(logger *slog.Logger) error {
	var cfg config

	flag.StringVar(&cfg.baseURL, "base-url", "http://localhost:9697", "base URL for the application")
	flag.IntVar(&cfg.httpPort, "http-port", 9697, "port to listen on for HTTP requests")
	flag.IntVar(&cfg.internalHttpPort, "internal-http-port", 9698, "port to listen on for internal API HTTP requests")
	flag.StringVar(&cfg.db.dsn, "db-dsn", "db.sqlite?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=on&_synchronous=NORMAL&_txlock_immediate", "sqlite3 DSN")
	flag.BoolVar(&cfg.db.automigrate, "db-automigrate", true, "run migrations on startup")

	showVersion := flag.Bool("version", false, "display version and exit")

	flag.Parse()

	if *showVersion {
		fmt.Printf("version: %s\n", version.Get())
		return nil
	}

	db, err := database.New(cfg.db.dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	if cfg.db.automigrate {
		err = db.MigrateUp()
		if err != nil {
			return err
		}
	}

	app := &application{
		config: cfg,
		db:     db,
		logger: logger,
	}

	return app.serveHTTP()
}
