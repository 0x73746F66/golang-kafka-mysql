SHELL := /bin/bash
-include .env
export $(shell sed 's/=.*//' .env)

.PHONY: help

help: ## This help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help

setup: ## Creates docker networks and volumes
	touch .env
	docker network create fiskil 2>/dev/null || true
	docker volume create --name=mysql-data 2>/dev/null || true
	docker volume create --name=kafka-data 2>/dev/null || true
	docker volume create --name=zookeeper-data 2>/dev/null || true

mysql-init: ## applies mysql schema and initial test data
	docker-compose \
		-f .development/docker-compose.yaml \
		--project-directory . \
		exec mysql bash -c "mysql -uroot -p'$(MYSQL_DEV_PASSWORD)' -q -s < /tmp/sql/schema.sql"
	docker-compose \
		-f .development/docker-compose.yaml \
		--project-directory . \
		exec mysql bash -c "mysql -uroot -p'$(MYSQL_DEV_PASSWORD)' -q -s < /tmp/sql/test_data.sql"

update: ## Pulls the latest go mysql zookeeper and kafka images
	docker-compose \
		-f .development/docker-compose.yaml \
		--project-directory . \
		pull mysql zookeeper kafka
	docker pull docker.io/library/golang:1.16-bullseye

up: ## Starts the publisher, kafka, mysql, and zookeeper containers
	docker-compose \
		-f .development/docker-compose.yaml \
		--project-directory . \
		up -d 

build:
	docker-compose \
		-f .development/docker-compose.yaml \
		--project-directory . \
		build publisher

buildnc:
	docker-compose \
		-f .development/docker-compose.yaml \
		--project-directory . \
		--no-cache \
		build publisher

run-publisher:
	docker-compose \
		-f .development/docker-compose.yaml \
		--project-directory . \
		up publisher

run: build run-publisher ## Build and run publisher

services: ## Starts all containers
	docker-compose \
		-f .development/docker-compose.yaml \
		--project-directory . \
		up -d mysql zookeeper kafka

ps: ## Shows this projects running docker containers
	docker-compose \
		-f .development/docker-compose.yaml \
		--project-directory . \
		ps

down: ## Bring down containers and removes anything else orphaned
	docker-compose  \
		-f .development/docker-compose.yaml \
		--project-directory . \
		down --remove-orphans

test: ## runs go test on local not in Docker
	docker-compose  \
		-f .development/docker-compose.yaml \
		--project-directory . \
		run publisher go test -v ./...

semgrep-sast-ci: ## run core semgrep rules for CI
	semgrep --disable-version-check -q --strict --error -o semgrep-ci.json --json --timeout=0 --config=p/r2c-ci --lang=py src/**/*.gp

sast: ## runs semgrep (install with `python3 -m pip install semgrep`)
	semgrep --strict --timeout=0 --config=p/r2c-ci --lang=py src/**/*.go
