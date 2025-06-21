# Build stage
FROM golang:1.24-alpine AS builder

# Install git and ca-certificates
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build arguments for version information
ARG VERSION=unknown
ARG GIT_COMMIT=unknown
ARG BUILD_DATE=unknown

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-X github.com/roman-povoroznyk/k6s/pkg/version.Version=${VERSION} -X github.com/roman-povoroznyk/k6s/pkg/version.GitCommit=${GIT_COMMIT} -X github.com/roman-povoroznyk/k6s/pkg/version.BuildDate=${BUILD_DATE}" \
    -a -installsuffix cgo -o k6s .

# Final stage - distroless
FROM gcr.io/distroless/static:nonroot

# Copy timezone data and certificates from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary
COPY --from=builder /app/k6s /usr/local/bin/k6s

# Use non-root user
USER nonroot:nonroot

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/usr/local/bin/k6s", "--version"]

# Default command
ENTRYPOINT ["/usr/local/bin/k6s"]
CMD ["server", "--port", "8080"]
