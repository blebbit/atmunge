.PHONY: all build up update down start-db status logs

all:
	go test -v ./...

.env:
	@cp example.env .env
	@echo "Please edit .env to suit your environment before proceeding"
	@exit 1

build: .env
	@docker compose build

up: .env
	@docker compose up -d --build

update: up

down:
	@docker compose down

start-db: .env
	@docker compose up -d postgres

status:
	@docker compose stats

logs:
	@docker compose logs -f -n 50

watch:
	watch -n 2 make watch.cmds
watch.cmds:
	@du -hd .
	@docker compose logs plc -n 2