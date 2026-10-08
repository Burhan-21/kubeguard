.PHONY: build test test-race lint cover clean docker-build helm-lint scan-self all

build:
	go build -o bin/kubeguard ./cmd/kubeguard

test:
	go test ./... -v

test-race:
	go test ./... -race -v

lint:
	gofmt -l . && go vet ./...

cover:
	go test ./... -coverprofile=coverage.out

clean:
	rm -rf bin/ coverage.out

docker-build:
	docker build -t kubeguard:latest .

helm-lint:
	helm lint charts/kubeguard

scan-self:
	bin/kubeguard scan charts/kubeguard/templates/

release-build:
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o bin/kubeguard-linux-amd64 ./cmd/kubeguard
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "-s -w" -o bin/kubeguard-linux-arm64 ./cmd/kubeguard
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o bin/kubeguard-darwin-amd64 ./cmd/kubeguard
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o bin/kubeguard-darwin-arm64 ./cmd/kubeguard
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o bin/kubeguard-windows-amd64.exe ./cmd/kubeguard

all: lint test build
