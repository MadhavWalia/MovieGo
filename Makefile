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


# ==================================================================================== #
# BUILD
# ==================================================================================== #


current_time = $(shell date --iso-8601=seconds)
git_description = $(shell git describe --always --dirty --tags --long)
linker_flags = '-s -X main.buildTime=${current_time} -X main.version=${git_description}'


## build/api: build the cmd/api application
.PHONY: build/api
build/api:
	@echo "Building the cmd/api..."
	go build -ldflags=${linker_flags} -o=./bin/api ./cmd/api
	GOOS=linux GOARCH=amd64 go build -ldflags=${linker_flags} -o=./bin/linux_amd64/api ./cmd/api