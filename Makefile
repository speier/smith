.PHONY: run test build clean

run:
	go run .

test:
	go test -v ./...

build:
	go build -o smith .

clean:
	rm -f smith
	go clean
