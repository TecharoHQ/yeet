.PHONY: build clean deps all lint test package

build:
	go build -o ./var/yeet ./cmd/yeet

clean:
	rm ./var/yeet* ||:

deps:
	go mod download

all: build lint test clean package

lint:
	go vet ./...
	go tool staticcheck ./...

test:
	go test ./...

package:
	go run ./cmd/yeet