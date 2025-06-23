# Build stage
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-X github.com/roman-povoroznyk/k8s/cmd.Version=${VERSION:-dev} \
              -X github.com/roman-povoroznyk/k8s/cmd.BuildTime=$(date -u '+%Y-%m-%d_%H:%M:%S') \
              -X github.com/roman-povoroznyk/k8s/cmd.CommitHash=${COMMIT_HASH:-unknown}" \
    -a -installsuffix cgo \
    -o k8s \
    ./main.go

# Final stage - Distroless
FROM gcr.io/distroless/static:nonroot

# Copy the binary from builder stage
COPY --from=builder /app/k8s /k8s

# Use non-root user
USER nonroot:nonroot

# Set entrypoint
ENTRYPOINT ["/k8s"]
