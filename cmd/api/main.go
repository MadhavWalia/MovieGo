package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"

	"moviego.madhav.net/internal/data"
	"moviego.madhav.net/internal/logs"
)


const version = "1.0.0"

type config struct {
	port int
	env string
	db struct {
		dsn string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime string
	}
	limiter struct {
		rps float64
		burst int
		enabled bool
	}
}


type application struct {
	config config
	logger *logs.Logger
	models data.Models
}


func main() {

	// Declare an instance of the config struct.
	var cfg config

	// Read the value dsn from the .env file, or use the default value
	dsn, err := loadEnv("MOVIEGO_DB_DSN")
	if err != nil {
		log.Fatal(err)
	}

	// Read the value of the port and env command-line flags into the config struct

	//Application Settings Flags
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", dsn, "PostgreSQL DSN")

	// Database Settings Flags
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

	// Rate Limiter Settings Flags
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Rate limiter enabled")


	// Parse the command-line flags
	flag.Parse()


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