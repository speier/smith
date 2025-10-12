.PHONY: build run test clean install release

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

# Auto-increment patch version and release
LATEST_TAG := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")

release:
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "❌ Working directory is dirty. Commit or stash changes first."; \
		exit 1; \
	fi
	@echo "Current version: $(LATEST_TAG)"
	$(eval NEXT_VERSION := $(shell echo $(LATEST_TAG) | awk -F. '{$$NF = $$NF + 1;} 1' | sed 's/ /./g'))
	@echo "Next version: $(NEXT_VERSION)"
	@read -p "Release $(NEXT_VERSION)? [y/N] " -n 1 -r && echo && \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		git tag -a $(NEXT_VERSION) -m "Release $(NEXT_VERSION)" && \
		git push origin $(NEXT_VERSION) && \
		echo "✓ Tagged and pushed $(NEXT_VERSION)"; \
	fi
