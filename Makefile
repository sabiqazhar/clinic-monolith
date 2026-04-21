include .env
export

MIGRATE=migrate

# ========================
# CREATE MIGRATION
# ========================
create-pg:
	$(MIGRATE) create -ext sql -dir migrations/postgres -seq $(name)

create-my:
	$(MIGRATE) create -ext sql -dir migrations/mysql -seq $(name)

# ========================
# POSTGRES
# ========================
pg-up:
	$(MIGRATE) -database "$(PG_URL)" -path migrations/postgres up

pg-down:
	$(MIGRATE) -database "$(PG_URL)" -path migrations/postgres down 1

pg-force:
	$(MIGRATE) -database "$(PG_URL)" -path migrations/postgres force $(version)

pg-version:
	$(MIGRATE) -database "$(PG_URL)" -path migrations/postgres version

# ========================
# MYSQL
# ========================
my-up:
	$(MIGRATE) -database "$(MYSQL_URL)" -path migrations/mysql up

my-down:
	$(MIGRATE) -database "$(MYSQL_URL)" -path migrations/mysql down 1

my-force:
	$(MIGRATE) -database "$(MYSQL_URL)" -path migrations/mysql force $(version)

my-version:
	$(MIGRATE) -database "$(MYSQL_URL)" -path migrations/mysql version:``

# ========================
# MODULE SKELETON
# ========================
create-module:
	@if [ -z "$(name)" ]; then \
		echo "❌ Usage: make create-module name=<module_name>"; \
		echo "   Example: make create-module name=billing"; \
		exit 1; \
	fi; \
	mkdir -p internal/modules/$(name)/domain && \
	mkdir -p internal/modules/$(name)/handler && \
	mkdir -p internal/modules/$(name)/repository/query && \
	mkdir -p internal/modules/$(name)/service && \
	echo "package domain" > internal/modules/$(name)/domain/interfaces.go && \
	echo "package handler" > internal/modules/$(name)/handler/$(name).go && \
	echo "package repository" > internal/modules/$(name)/repository/$(name).go && \
	echo " -- creating query in this file. example: '-- name: FindItemByID :one' then add query bellow" > internal/modules/$(name)/repository/queries.sql && \
	echo "package service" > internal/modules/$(name)/service/$(name).go && \
	echo "package $(name)" > internal/modules/$(name)/provider.go && \
	echo "" && \
	echo "Module '$(name)' created:" && \
	find internal/modules/$(name) -type f
