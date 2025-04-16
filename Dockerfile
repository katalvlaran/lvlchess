# Dockerfile for the Go backend (Telegram Bot)

# ======================================
# 1) Builder Stage
#    - We compile the Go binary in a container with the necessary build tools
# ======================================
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum first to leverage caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project (lvlchess) into the container
COPY . .

# Build the binary (bot.go as main entry)
RUN go build -o lvlchess ./cmd/bot.go

# ======================================
# 2) Final Minimal Stage
#    - Use a lightweight Alpine image to run the compiled binary
# ======================================
FROM alpine:3.17
WORKDIR /app

# Copy the compiled binary from builder
COPY --from=builder /app/lvlchess /app/

# Expose port 8080 in case we want HTTP endpoints
EXPOSE 8080

# Default command: run the lvlchess binary (the Telegram bot)
CMD ["/app/lvlchess"]
