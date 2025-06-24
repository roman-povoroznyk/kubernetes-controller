APP = k8s-ctrl
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_FLAGS = -v -o $(APP) -ldflags "-X=github.com/roman-povoroznyk/kubernetes-controller/cmd.appVersion=$(VERSION)"

.PHONY: all build test run docker-build clean

all: build

build:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(BUILD_FLAGS) main.go

test:
	go test ./...

coverage:
  go test -coverprofile=coverage.out ./...
  go tool cover -html=coverage.out

run:
	go run main.go

docker-build:
	docker build --build-arg VERSION=$(VERSION) -t $(APP):latest .

clean:
	rm -f $(APP)
