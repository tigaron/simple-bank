include .env

postgres:
	docker run --name postgres --hostname postgres --publish ${POSTGRES_PORT}:${POSTGRES_PORT} --user ${shell id -u} --volume ${shell pwd}/db/data:/var/lib/postgresql/data --env POSTGRES_USER=${POSTGRES_USER} --env POSTGRES_PASSWORD=${POSTGRES_PASSWORD} --restart always --detach postgres:alpine

createdb:
	docker exec --interactive --tty postgres createdb --username=${POSTGRES_USER} --owner=${POSTGRES_USER} simple_bank

dropdb:
	docker exec --interactive --tty postgres dropdb simple_bank

initschema:
	migrate create -ext sql -dir db/migration -seq init_schema

migrateup:
	migrate -path db/migration -database "postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/simple_bank?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/simple_bank?sslmode=disable" -verbose down

sqlc:
	docker run --rm --volume ${shell pwd}:/src --workdir /src kjconroy/sqlc generate

test:
	go test -v -cover ./...

.PHONY: postgres createdb dropdb initschema migrateup migratedown sqlc test
