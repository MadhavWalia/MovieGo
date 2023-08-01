package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"

	"moviego.madhav.net/internal/data"
	"moviego.madhav.net/internal/logs"
	"moviego.madhav.net/internal/mail"
)

var (
	version   string
	buildTime string
)

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
	cors struct {
		trustedOrigins []string
	}
}

type application struct {
	config config
	logger *logs.Logger
	models data.Models
	mailer mail.Mailer
	wg     sync.WaitGroup
}

func main() {

	// Declare an instance of the config struct.
	var cfg config

	// Read the value of the port and env command-line flags into the config struct

	//Application Settings Flags
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", "", "PostgreSQL DSN")

	// Database Settings Flags
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 2525, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

	// Rate Limiter Settings Flags
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Rate limiter enabled")

	// SMTP Settings Flags
	flag.StringVar(&cfg.smtp.host, "smtp-host", "smtp.mailtrap.io", "SMTP server hostname")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 25, "SMTP server port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "", "SMTP server username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "", "SMTP server password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "", "SMTP sender email address")

	// CORS Settings Flags
	flag.Func("cors-trusted-origins", "CORS trusted origins (space separated)", func(val string) error {
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})

	// Version Flag
	displayVersion := flag.Bool("version", false, "Display version and exit")

	// Parse the command-line flags
	flag.Parse()

	// If the version flag is provided, print the version and exit
	if *displayVersion {
		fmt.Printf("version:\t%s\n", version)
		fmt.Printf("build time:\t%s\n", buildTime)
		os.Exit(0)
	}

	// Initialize a new logger which writes messages to the standard outstream
	logger := logs.New(os.Stdout, logs.LevelInfo)

	// Initialize a new connection pool, passing in the DSN from the config struct
	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	defer db.Close()

	// Log a message to say that the connection pool has been successfully
	logger.PrintInfo("database connection pool established", nil)

	// Initialize a new instance of application containing the dependencies
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mail.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

	// Start the HTTP server
	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}
}

// The openDB() function wraps sql.Open() and returns a sql.DB connection pool
func openDB(cfg config) (*sql.DB, error) {

	// Use sql.Open() to create an empty connection pool, using the DSN from the config struct
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	// Set the maximum number of open (in-use + idle) connections in the pool.
	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	// Parse the dbMaxIdleTime setting from the config struct
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}

	// Set the maximum idle timeout.
	db.SetConnMaxIdleTime(duration)

	// Set a 5 second deadline context for a deadline on connection attempts
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use PingContext() to establish a new connection to the database, passing in the context
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	// Return the sql.DB connection pool
	return db, nil
}
