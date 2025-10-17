.PHONY: build run test clean

build:
	go build -o smith .

run:
	go run .

test:
	go test -v ./...

clean:
	rm -f smith
	go clean
