# vars
DB_URL=postgres://atmirror:atmirror@localhost:5555/atmirror?sslmode=disable
DATE=$(shell date +"%Y%m%d")

.PHONY: all build up update down start-db status logs

all:
	go test -v ./...

.env:
	@cp env-example .env
	@echo "Please edit .env to suit your environment before proceeding"
	@exit 1

build:
	@docker compose build

up:
	@docker compose up -d --build

update: up

down:
	@docker compose down

start-db:
	@docker compose up -d postgres

psql:
	docker exec -it at-mirror-db-1 psql -U atmirror -d atmirror

status:
	@docker compose stats

logs:
	@docker compose logs -f -n 50

watch:
	watch -n 2 make watch.cmds
watch.cmds:
	@du -hd .
	@docker compose logs plc -n 2


dump.plc_log_entries.raw:
	pg_dump -d "$(DB_URL)" -t plc_log_entries -Z zstd -f plc_log_entries-raw-$(DATE).sql.zst
restore.plc_log_entries.raw:
	pg_restore -d "$(DB_URL)" -t plc_log_entries -f plc_log_entries-raw-$(DATE).sql.zst
upload.plc_log_entries.raw:
	rclone copy -P plc_log_entries-raw-$(DATE).sql.zst blebbit-public-bucket:public/plc

dump.pds_repos:
	pg_dump -d "$(DB_URL)" -t pds_repos -Z zstd -f pds_repos-$(DATE).sql.zst
restore.pds_repos:
	pg_restore -d "$(DB_URL)" -t pds_repos -f pds_repos-$(DATE).sql.zst
upload.pds_repos:
	rclone copy -P pds_repos-$(DATE).sql.zst blebbit-public-bucket:public/plc

dump.account_infos:
	pg_dump -d "$(DB_URL)" -t account_infos -Z zstd -f account_infos-$(DATE).sql.zst
restore.account_infos:
	pg_restore -d "$(DB_URL)" -t account_infos -f account_infos-$(DATE).sql.zst
upload.account_infos:
	rclone copy -P account_infos-$(DATE).sql.zst blebbit-public-bucket:public/plc