# Build Stage
FROM golang:1.22-alpine AS builder

LABEL stage=builder

WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker cache for dependencies
COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify

# Copy the rest of the application source code
COPY . .

# Build the application
# Using -a to force rebuilding of packages that are already up-to-date.
# Using -installsuffix cgo to prevent conflicts with host C libraries.
# Outputting to a specific path for easy copying in the next stage.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o /app/finance-tracker-app ./cmd/server/main.go

# Final Stage
FROM alpine:latest

# Install CA certificates for HTTPS calls (e.g., to OpenRouter AI)
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the compiled application binary from the build stage
COPY --from=builder /app/finance-tracker-app .

# Copy migrations directory.
# While GORM's AutoMigrate is used, having these files can be useful for reference
# or if a manual migration step is ever needed with these files.
COPY migrations ./migrations

# Copy web assets
COPY web/templates ./web/templates
COPY web/static ./web/static

# Expose the port the application runs on
EXPOSE 8080

# Set environment variables
# GIN_MODE=release is good for production to reduce logging and improve performance.
# Can be overridden in docker-compose.yml for development.
ENV GIN_MODE=release
ENV PORT=8080

# Command to run the application
ENTRYPOINT ["/app/finance-tracker-app"]
