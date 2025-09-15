postgres:
	docker run --name postgres15 -p 5433:5432 -e POSTGRES_USER=isha -e POSTGRES_PASSWORD=secret -d postgres:15-alpine

createdb:
	docker exec -it postgres15 createdb --username=isha --owner=isha usersdb

dropdb:
	docker exec -it postgres15 dropdb usersdb

migrateup:
	migrate -path db/migration -database "postgresql://isha:secret@localhost:5433/usersdb?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://isha:secret@localhost:5433/usersdb?sslmode=disable" -verbose down

test:
	go test -v -cover ./...

.PHONY: postgres createdb dropdb migrateup migratedown test
