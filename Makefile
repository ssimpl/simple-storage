GOUTIL=docker build -t goutil -f ./Dockerfile.util . && \
	docker run --rm \
	-e CGO_ENABLED=1 \
	-v "$(shell pwd):/src" \
	-w /src \
	goutil

.PHONY: generate
generate:
	$(GOUTIL) sh -c 'go generate -v ./...'

.PHONY: lint
lint:
	docker run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:v1.61 golangci-lint run -v

.PHONY: test
test:
	go test -timeout 1m -race -cover ./...

.PHONY: up
up:
	docker-compose up --build -d || docker compose up --build -d

.PHONY: down
down:
	docker-compose down || docker compose down

.PHONY: client
client:
	go run ./cmd/client/...
