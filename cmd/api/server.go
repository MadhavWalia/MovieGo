package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error {
	// Declare a HTTP server with necessary settings
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		ErrorLog:     log.New(app.logger, "", 0),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Creating a shutdownError channel to carry error values given by the server.Shutdown() method
	shutdownError := make(chan error)

	// Background goroutine to gracefully shutdown the server when the shutdown signal is received
	go func() {
		// Creating a quit channel which carries os.Signal values
		quit := make(chan os.Signal, 1)

		// Using signal.Notify() to listen for incoming SIGINT and SIGTERM signals
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		// Reading the signal from the channel. This code will block until a signal is received
		s := <-quit

		// Logging a message to say that the signal has been caught
		app.logger.PrintInfo("caught the signal", map[string]string{
			"signal": s.String(),
		})

		// Creating a context with a 5-second timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Calling Shutdown() method on our server, passing in the context and returning the error only id we get an error
		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		// Logging a message to say that we're waiting for any background goroutines to complete their tasks
		app.logger.PrintInfo("completing background tasks", map[string]string{
			"addr": srv.Addr,
		})

		// Blocking until the all the background goroutines have completed
		app.wg.Wait()

		// Returning nil to indicate that the shutdown completed successfully
		shutdownError <- nil
	}()

	// Log a message to say that the server is starting
	app.logger.PrintInfo("starting server", map[string]string{
		"addr": srv.Addr,
		"env":  app.config.env,
	})

	// Calling the ListenAndServe() method on our HTTP server
	err := srv.ListenAndServe()
	// If the return error is that the server has been closed, it means that the server has been shut down gracefully
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// Any other error indicates an unexpected failure in graceful shutdown
	err = <-shutdownError
	if err != nil {
		return err
	}

	// Log a message to say that the server has stopped
	app.logger.PrintInfo("stopped server", map[string]string{
		"addr": srv.Addr,
	})

	return nil
}
