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

psql:
	docker exec -it at-mirror-postgres-1 psql -U postgres -d plc

status:
	@docker compose stats

logs:
	@docker compose logs -f -n 50

watch:
	watch -n 2 make watch.cmds
watch.cmds:
	@du -hd .
	@docker compose logs plc -n 2

DB_URL=postgres://atmirror:atmirror@localhost:5432/atmirror?sslmode=disable
DATE=$(shell date +"%Y%m%d")
dump.plc_log_entries.raw:
	pg_dump -d "$(DB_URL)" -t plc_log_entries -Z zstd -f plc_log_entries-raw-$(DATE).sql.zst
restore.plc_log_entries.raw:
	pg_restore -d "$(DB_URL)" -t plc_log_entries -f plc_log_entries-raw-$(DATE).sql.zst
upload.plc_log_entries.raw:
	rclone copy plc_log_entries-raw-$(shell date +"%Y%m%d").sql.zst blebbit-public-bucket:public/plc