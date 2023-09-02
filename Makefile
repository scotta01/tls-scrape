# Makefile for TLS-Scrape

# Variables
BINARY_NAME = tls-scrape
DOCKER_IMAGE_NAME = scotta01/tls-scrape
GO_FILES = $(shell find . -name "*.go")
TEST_FQDN = www.google.com

# Compile the application
build:
	@echo "Building the Go application..."
	go build -o $(BINARY_NAME) cmd/tls-scrape/main.go

# Run the application
run: build
	@echo "Running the application..."
	./$(BINARY_NAME) --fqdn $(TEST_FQDN)

# Create a Docker image
docker-build: Dockerfile $(GO_FILES)
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE_NAME) .

# Clean up
clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME)
	docker rmi $(DOCKER_IMAGE_NAME)

.PHONY: build run docker-build clean
