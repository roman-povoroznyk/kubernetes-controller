# syntax=docker/dockerfile:1.4
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -v -o k8s-ctrl -ldflags "-X=github.com/roman-povoroznyk/kubernetes-controller/cmd.appVersion=$VERSION" main.go

# Final stage
FROM gcr.io/distroless/static-debian12
WORKDIR /
COPY --from=builder /app/k8s-ctrl .
EXPOSE 8080
ENTRYPOINT ["/k8s-ctrl"]
CMD ["server", "--server-port", "8080"]
