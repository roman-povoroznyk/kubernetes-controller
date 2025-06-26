APP = k8s-ctrl
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_FLAGS = -v -o $(APP) -ldflags "-X=github.com/roman-povoroznyk/kubernetes-controller/cmd.appVersion=$(VERSION)"

# Tool Versions
ENVTEST_VERSION ?= release-0.19
LOCALBIN ?= $(shell pwd)/bin

# Tool Binaries
ENVTEST ?= $(LOCALBIN)/setup-envtest

.PHONY: all build test test-coverage envtest format lint run docker-build clean

all: build

## Location to install dependencies to
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Download setup-envtest locally if necessary
envtest: $(ENVTEST)
$(ENVTEST): $(LOCALBIN)
	$(call go-install-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest,$(ENVTEST_VERSION))

format:
	gofmt -s -w ./

lint:
	golangci-lint run

build:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(BUILD_FLAGS) main.go

test: envtest
	go install gotest.tools/gotestsum@latest
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use --bin-dir $(LOCALBIN) -p path)" gotestsum --junitfile report.xml --format testname ./...

test-coverage: envtest
	go install github.com/boumenot/gocover-cobertura@latest
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use --bin-dir $(LOCALBIN) -p path)" go test -coverprofile=coverage.out -covermode=count ./...
	go tool cover -func=coverage.out
	gocover-cobertura < coverage.out > coverage.xml

# Quick local coverage check - opens HTML report in browser
coverage: envtest
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use --bin-dir $(LOCALBIN) -p path)" go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

run:
	go run main.go

docker-build:
	docker build --build-arg VERSION=$(VERSION) -t $(APP):latest .

clean:
	rm -f $(APP) coverage.out coverage.xml report.xml
	rm -rf $(LOCALBIN)

# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f "$(1)-$(3)" ] || { \
	set -e; \
	package=$(2)@$(3) ;\
	echo "Downloading $${package}" ;\
	rm -f $(1) || true ;\
	GOBIN=$(LOCALBIN) go install $${package} ;\
	mv $(1) $(1)-$(3) ;\
} ;\
ln -sf $(1)-$(3) $(1)
endef
