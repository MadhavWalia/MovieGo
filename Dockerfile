# Using the golang 1.20.6 as base image
FROM golang:1.20.6 as builder

# Install git
RUN apt-get update && apt-get install -y git

# Install make
RUN apt-get update && apt-get install -y make

# Install the migrate cli tool for database migrations
# RUN  curl -L https://github.com/golang-migrate/migrate/releases/download/v4.14.1/migrate.linux-amd64.tar.gz | tar xvz
# RUN mv migrate.linux-amd64 $GOPATH/bin/migrate

# Install staticcheck for linting
RUN go install honnef.co/go/tools/cmd/staticcheck@latest

# Set the working directory to the app directory
WORKDIR /app

# Copy the current directory contents into the container at /app
COPY . .

# Make the binary directory
RUN mkdir -p bin

# Audit the code
RUN make audit

# Build the app
RUN make build/api


## Stage 2: Run the app

# Using alpine as base image
FROM alpine:latest

# Install make
RUN apk add --no-cache ca-certificates
RUN apk add --no-cache make

# Set the working directory to the app directory
WORKDIR /app

# apk add gcompat for the binary to run on alpine
RUN apk add libc6-compat

# Copy the binary from the builder stage
COPY --from=builder /app/bin/ /app/bin/
RUN chmod +x /app/bin/api

# Copy the makefile and .env from the builder stage
COPY --from=builder /app/Makefile /app/Makefile
COPY --from=builder /app/.env /app/.env

# Expose port 4000 to the outside world
EXPOSE 4000

# Run the executable
CMD ["make","run/binary"]




