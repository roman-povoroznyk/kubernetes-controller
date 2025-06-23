# Build stage
FROM golang:1.24-alpine AS builder

# Install git and ca-certificates
RUN apk --no-cache add git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o k6s .

# Final stage - distroless image
FROM gcr.io/distroless/static:nonroot

# Copy ca-certificates from builder stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary from builder stage
COPY --from=builder /app/k6s /usr/local/bin/k6s

# Use non-root user
USER nonroot:nonroot

# Expose port
EXPOSE 8080

# Run the binary
ENTRYPOINT ["/usr/local/bin/k6s"]
