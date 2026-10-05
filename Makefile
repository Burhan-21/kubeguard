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

all: lint test build
