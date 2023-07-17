package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)


const version = "1.0.0"

type config struct {
	port int
	env string
	db struct {
		dsn string
	}
}

type application struct {
	config config
	logger *log.Logger
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
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", dsn, "PostgreSQL DSN")
	flag.Parse()

	// Initialize a new logger which writes messages to the standard outstream
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// Initialize a new connection pool, passing in the DSN from the config struct
	db, err := openDB(cfg)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	// Log a message to say that the connection pool has been successfully
	logger.Printf("database connection pool established")

	// Initialize a new instance of application containing the dependencies
	app := &application{
		config: cfg,
		logger: logger,
	}

	// Declare a HTTP server with some timeout settings, and bind the servemux
	srv := &http.Server{
		Addr: fmt.Sprintf(":%d", cfg.port),
		Handler: app.routes(),
		IdleTimeout: time.Minute,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Start the HTTP server
	logger.Printf("starting %s server on %s", cfg.env, srv.Addr)
	err = srv.ListenAndServe()
	logger.Fatal(err)
}

// The openDB() function wraps sql.Open() and returns a sql.DB connection pool
func openDB(cfg config) (*sql.DB, error) {
	
	// Use sql.Open() to create an empty connection pool, using the DSN from the config struct
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

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