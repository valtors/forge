.PHONY: build test lint clean coverage

build:
	go build -o forge .

test:
	go test ./... -count=1

coverage:
	go test ./... -count=1 -coverprofile=coverage.out
	go tool cover -func=coverage.out | grep total

lint:
	gofmt -l .
	go vet ./...

clean:
	rm -f forge coverage.out
