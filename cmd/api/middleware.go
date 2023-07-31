package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
	"moviego.madhav.net/internal/data"
	"moviego.madhav.net/internal/validator"
)

// Middleware for panic recovery
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a deferred function which will always be run at panic as the stack unwinds
		defer func() {
			if err := recover(); err != nil {
				// Set a "Connection: close" header on the response
				w.Header().Set("Connection", "close")

				// Call the app.serverErrorResponse() method to send a 500 Internal Server
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		// Call the next handler in the chain
		next.ServeHTTP(w, r)
	})
}


// Middleware for rate limiting
func (app *application) rateLimit(next http.Handler) http.Handler {
	// Declare a client struct to hold the rate limiter and last seen time for each client
	type client struct {
		limiter *rate.Limiter
		lastSeen time.Time
	}

	// Declare a mutex and a map to hold the rate limiters for each IP address
	var (
		mu sync.Mutex
		clients = make(map[string]*client)
	)


	// Launch a background goroutine which removes old entries from the clients map once every minute
	go func() {
		for {
			time.Sleep(time.Minute)

			// Lock the mutex to prevent any rate limiter checks from happening while the cleanup is taking place
			mu.Lock()

			// Loop through all clients. If they haven't been seen within the last three minutes, delete the corresponding entry from the map
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3 * time.Minute {
					delete(clients, ip)
				}
			}

			// Unlock the mutex
			mu.Unlock()
		}
	}()


	// Return a closure over the limiter
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		//Check if rate limiting is enabled
		if !app.config.limiter.enabled {
			// Extracting the client's IP address from the request
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}


			// Locking the mutex to prevent this code from being executed concurrently
			mu.Lock()


			// Checking to see if the IP address already exists in the map, initializing one if not
			if _, ok := clients[ip]; !ok {
				clients[ip] = &client{
					limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst),
				}
			}
			// Updating the last seen time for the client
			clients[ip].lastSeen = time.Now()


			// Checking whether the limiter is allowing the request. If not, return a 429
			if !clients[ip].limiter.Allow() {
				// Unlock the mutex and return a 429 Too Many Requests response
				mu.Unlock()
				app.rateLimitExceededResponse(w, r)
				return
			}

			// Unlocking the mutex before calling the next handler in the chain
			mu.Unlock()
		}

		// Calling the next handler in the chain
		next.ServeHTTP(w, r)
	})
}


// Middleware for authentication
func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Adding the "Vary: Authorization" header to the response
		w.Header().Add("Vary", "Authorization")


		// Extracting the value of the Authorization header from the request
		authorizationHeader := r.Header.Get("Authorization")


		// If there is no Authorization header found, use the contextSetUser() method to set the AnonymousUser in the request context and call the next handler in the chain and return
		if authorizationHeader == "" {
			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}


		// If the header is found, then extract the token from the header
		headerParts := strings.Split(authorizationHeader, " ")
		// Checking if the header is in the correct format (Bearer <token>)
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}


		// Retrieving the token from the headerParts and performing validation
		token := headerParts[1]
		v := validator.New()
		if data.ValidateTokenPlaintext(v, token); !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}


		// Retrieving the details of the user from the token
		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
				case errors.Is(err, data.ErrRecordNotFound):
					app.invalidAuthenticationTokenResponse(w, r)
				default:
					app.serverErrorResponse(w, r, err)
			}
			return
		}


		// Adding the user details to the request context
		r = app.contextSetUser(r, user)


		// Calling the next handler in the chain
		next.ServeHTTP(w, r)
	})
}


// Middleware for requiring an authenticated user
func (app *application) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Retrieving the user from the request context
		user := app.contextGetUser(r)


		// If the user is anonymous, return a 401 Unauthorized response
		if user.IsAnonymous() {
			app.authenticationRequiredResponse(w, r)
			return
		}


		// Calling the next handler in the chain
		next.ServeHTTP(w, r)
	}) 
}


// Middleware for requiring an activated user
func (app *application) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Retrieving the user from the request context
		user := app.contextGetUser(r)


		// If the user is not activated, return a 403 Forbidden response
		if !user.Activated {
			app.inactiveAccountResponse(w, r)
			return
		}


		// Calling the next handler in the chain
		next.ServeHTTP(w, r)
	})


	// Wrap the middleware around the requireAuthenticatedUser() middleware
	return app.requireAuthenticatedUser(fn)
}