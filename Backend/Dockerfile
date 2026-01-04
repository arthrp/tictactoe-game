# == Build stage ==
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy dependency files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=0 ensures a static binary for the final lightweight image
RUN CGO_ENABLED=0 GOOS=linux go build -o tictactoe-backend .


# == Runtime stage ==
FROM alpine:latest

# Install ca-certificates in case the app needs to make HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/tictactoe-backend .

# Expose the port the application listens on
EXPOSE 8080

# Run the binary
CMD ["./tictactoe-backend"]

