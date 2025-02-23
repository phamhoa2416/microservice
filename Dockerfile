# Stage 1: Build the Golang application
FROM golang:1.23-alpine3.20 AS builder

WORKDIR /app

# Copy Go modules and install dependencies
COPY go.mod go.sum ./
RUN go mod tidy

# Copy the application source code
COPY . .

# Build the application
RUN go build -o main

# Stage 2: Run Redis and the Go application
FROM redis:latest

WORKDIR /app

EXPOSE 6379

COPY --from=builder /app/main .

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
CMD ["/entrypoint.sh"]
