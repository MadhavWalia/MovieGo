# Using the golang 1.20.6 as base image
FROM golang:1.20.6

# Install git
RUN apt-get update && apt-get install -y git

# Install make
RUN apt-get update && apt-get install -y make

# Install the migrate cli tool for database migrations
RUN  curl -L https://github.com/golang-migrate/migrate/releases/download/v4.14.1/migrate.linux-amd64.tar.gz | tar xvz
RUN mv migrate.linux-amd64 $GOPATH/bin/migrate

# Install staticcheck for linting
RUN go install honnef.co/go/tools/cmd/staticcheck@latest

# Set the working directory to the app directory
WORKDIR /app

# Copy the current directory contents into the container at /app
COPY . .

# Expose the port 4000
EXPOSE 4000

# Run the app
CMD ["make", "run/api"]


