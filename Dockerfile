FROM golang:bookworm AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Install gcc and libc-dev for CGO (required by go-sqlite3)
RUN apt-get update && apt-get install -y gcc libc-dev

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app (Enable CGO for sqlite3)
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o bot_app ./cmd/app/main.go

# Start a new stage from a smaller base image (using bookworm-slim for compatibility with the glibc version used in builder)
FROM debian:bookworm-slim

WORKDIR /app

# Install CA certificates for HTTPS (Google Sheets API)
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/bot_app .
COPY --from=builder /app/.env* ./
# Kita meng-copy credentials.json (jika ada) saat build
COPY --from=builder /app/credentials.json* ./

# Command to run the executable
CMD ["./bot_app"]
