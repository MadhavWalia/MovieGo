include .env

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

#help: prints this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'


.PHONY: confirm
confirm:
	@echo "Are you sure you want to run this command? [y/N] " && read ans && [ $${ans:-N} = y ]


# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #


## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	@go run ./cmd/api -db-dsn=${MOVIEGO_DB_DSN} -smtp-sender=${SMTP_SENDER} -smtp-username=${SMTP_USERNAME} -smtp-password=${SMTP_PASSWORD}


## db/migrations/new name=$1: create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo "Creating a new migration file for $(name)..."
	@migrate create -ext sql -dir ./migrations -seq $(name)


## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo "Running the up migrations..."
	@migrate -path ./migrations -database "postgres://madhav:pa55word@localhost:5432/postgres?sslmode=disable" -verbose up


# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #


## audit: tidy dependencies and format, vet and test all code
.PHONY: audit
audit:
	@echo "Tidying and verifying dependencies..."
	go mod tidy
	go mod verify

	@echo "Formatting code..."
	go fmt ./...

	@echo "Vetting code..."
	go vet ./...
	staticcheck ./...

	@echo "Running tests..."
	go test -race -vet=off ./...
