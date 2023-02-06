include app.env

postgres:
	docker run --name postgres --hostname postgres --publish ${POSTGRES_PORT}:${POSTGRES_PORT} --user ${shell id -u} --volume ${shell pwd}/db/data:/var/lib/postgresql/data --env POSTGRES_USER=${POSTGRES_USER} --env POSTGRES_PASSWORD=${POSTGRES_PASSWORD} --restart always --detach postgres:alpine

createdb:
	docker exec --interactive --tty postgres createdb --username=${POSTGRES_USER} --owner=${POSTGRES_USER} ${POSTGRES_DB}

dropdb:
	docker exec --interactive --tty postgres dropdb ${POSTGRES_DB}

initschema:
	migrate create -ext sql -dir db/migration -seq init_schema

migrateup:
	migrate -path db/migration -database "${DB_SOURCE}" -verbose up

migratedown:
	migrate -path db/migration -database "${DB_SOURCE}" -verbose down

sqlc:
	docker run --rm --volume ${shell pwd}:/src --workdir /src kjconroy/sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/tigaron/simple-bank/db/sqlc Store

.PHONY: postgres createdb dropdb initschema migrateup migratedown sqlc test server mock
