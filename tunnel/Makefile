GO        ?= go
CGO       ?= 0
PKGS      := ./...
CMDS      := ./cmd/...

.PHONY: all build release debug test bench lint cover clean fmt vet setup run

# all runs vet and builds both binaries with CGO disabled.
all: vet build

# setup ensures module dependencies are downloaded and go.sum is populated.
setup:
	$(GO) mod tidy

# build compiles all commands with CGO disabled and trimpath for reproducible builds.
build:
	CGO_ENABLED=$(CGO) $(GO) build -trimpath $(CMDS)

# release builds with size-optimized flags and CGO disabled.
release:
	CGO_ENABLED=$(CGO) $(GO) build -trimpath -ldflags="-s -w" $(CMDS)

# debug builds with the race detector; CGO may be enabled implicitly here.
debug:
	$(GO) build -race $(CMDS)

# test runs the full test suite with race detector and without test caching.
test:
	$(GO) test -race -count=1 $(PKGS)

# bench runs benchmarks with memory statistics.
bench:
	$(GO) test -bench=. -benchmem $(PKGS)

# lint runs static analysis tools; assumes they are installed in PATH.
lint:
	golangci-lint run ./...
	staticcheck ./...

# cover generates an HTML coverage report.
cover:
	$(GO) test -coverprofile=coverage.out $(PKGS)
	$(GO) tool cover -html=coverage.out

# fmt formats all Go source files.
fmt:
	gofmt -w .

# vet runs go vet on all packages.
vet:
	$(GO) vet $(PKGS)

# run is a convenience target for starting the server and client in development;
# real wiring will be added by later agents.
run: build
	@echo "Binaries built. Start cmd/tunnel-server and cmd/tunnel-client as needed."

# clean removes build artifacts.
clean:
	rm -f coverage.out
	rm -f cmd/tunnel-server/tunnel-server
	rm -f cmd/tunnel-client/tunnel-client

