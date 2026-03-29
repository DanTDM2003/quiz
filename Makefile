SHELL := /bin/sh
COMPOSE := docker compose

.PHONY: help test build run clean docker-build up down logs

help:
	@echo "make test         - go test ./..."
	@echo "make build        - go build to bin/server"
	@echo "make run          - go run ./cmd/server"
	@echo "make docker-build - docker build -t quiz:latest ."
	@echo "make up           - docker compose up --build -d"
	@echo "make down         - docker compose down"
	@echo "make logs         - docker compose logs -f"

test:
	go test ./...

build:
	mkdir -p bin
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server

clean:
	rm -rf bin

docker-build:
	docker build -t quiz:latest .

up:
	$(COMPOSE) up --build -d

down:
	$(COMPOSE) down

logs:
	$(COMPOSE) logs -f
