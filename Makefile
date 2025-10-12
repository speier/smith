.PHONY: build run test clean install

build:
	go build -o smith .

install:
	go install .

run:
	go run .

test:
	go test -v ./...

clean:
	rm -f smith
	go clean

# Development commands
dev-orchestrate:
	go run . orchestrate --dry-run

dev-status:
	go run . status

# Example: run a single agent
dev-agent:
	go run . agent --role=implementation --task=1
