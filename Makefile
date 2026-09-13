VERSION ?= $(shell git describe --tags --always --dirty || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD || echo "unknown")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS = -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

.PHONY: build build-all test test-integration lint fmt security clean install uninstall

build:
	@mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o bin/ngvpn ./cmd/ngvpn

build-all:
	@mkdir -p bin/linux-amd64 bin/linux-arm64 bin/darwin-amd64 bin/darwin-arm64 bin/windows-amd64
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/linux-amd64/ngvpn ./cmd/ngvpn
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/linux-arm64/ngvpn ./cmd/ngvpn
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/darwin-amd64/ngvpn ./cmd/ngvpn
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/darwin-arm64/ngvpn ./cmd/ngvpn
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/windows-amd64/ngvpn.exe ./cmd/ngvpn

test:
	go test -v -cover -coverprofile=coverage.out ./...

test-integration:
	go test -v -tags=integration -coverprofile=coverage-integration.out ./...

lint:
	golangci-lint run ./...

fmt:
	gofmt -s -w .
	goimports -w .

security:
	govulncheck ./...
	gosec ./...

clean:
	rm -rf bin/ dist/ coverage.out coverage-integration.out

install: build
	install -d /usr/local/bin
	install -m 755 bin/ngvpn /usr/local/bin/ngvpn

uninstall:
	rm -f /usr/local/bin/ngvpn
