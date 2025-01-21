.PHONY: docker-build docker-run docker-dev

# Docker settings
IMAGE_NAME = nfl-scores
PORT = 2112
GIT_HASH = $(shell git rev-parse --short=6 HEAD)
BUILD_TIME = $(shell date +%Y%m%d-%H%M%S)
HOSTNAME = $(shell hostname)

# Build the Docker image
docker-build:
	docker build -t $(IMAGE_NAME):$(GIT_HASH) \
		-t $(IMAGE_NAME):$(GIT_HASH)-$(BUILD_TIME)-$(HOSTNAME) \
		-t $(IMAGE_NAME):latest .

# Run the production container
docker-run: docker-build
	docker run -p $(PORT):$(PORT) \
		-v $(PWD)/tmp/nats-store:/tmp/nats-store \
		-v /etc/localtime:/etc/localtime:ro \
		$(IMAGE_NAME):$(GIT_HASH)

# Run with source mounted for development
docker-dev:
	docker run -it --rm \
		-v $(PWD):/app \
		-v $(PWD)/tmp/nats-store:/tmp/nats-store \
		-v /etc/localtime:/etc/localtime:ro \
		-w /app \
		-p $(PORT):$(PORT) \
		golang:1.22-alpine \
		go run .
