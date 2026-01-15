.PHONY: test docker-build docker-up docker-down

test:
	go test ./... -v

docker-build:
	docker build -t afrogo-service-rate-engine-api:local .

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down --volumes
