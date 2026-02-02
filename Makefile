APP_NAME := k8s-jwks-proxy-amd64


.PHONY: all build clean

all: build

# Build the Go binary
build:
	@echo "Building binary..."
	go build -o $(APP_NAME) main.go

# Clean build artifacts and packages
clean:
	@echo "Cleaning up..."
	rm -f $(APP_NAME)

release:
	bin/release.sh

.PHONY: show-version
show-version:
	@echo "Current chart version: $(CURRENT_VERSION)"