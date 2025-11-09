# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies (git for go mod download)
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=0 creates a static binary (no CGO dependencies)
# -ldflags="-w -s" strips debug info and symbol table, reduces binary size
# -trimpath removes file system paths from the resulting executable
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-w -s" -o app ./cmd/main.go

# Final stage - minimal runtime image using scratch (empty base image)
# Note: Certificates are NOT needed because:
# - HTTP serving doesn't require certificates
# - PostgreSQL connection uses TCP (sslmode=disable in default config)
# - No HTTPS outbound API calls are made
# If you need HTTPS outbound calls in the future, add:
# COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
FROM scratch

# Copy the binary from builder stage
COPY --from=builder /build/app /app

# Expose port (Gin default is 8080)
EXPOSE 8080

# Run the application
ENTRYPOINT ["/app"]

