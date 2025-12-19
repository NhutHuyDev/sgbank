network:
	docker network create sgbank-network

postgres:
	docker run --name sgbankpg12 --network sgbank-network -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:12-alpine

createdb: 
	docker exec -it sgbankpg12 createdb --username=root --owner=root sgbank

dropdb:
	docker exec -it sgbankpg12 dropdb sgbank

migrateup:
	migrate -path db/migration -database "postgres://root:secret@127.0.0.1:5432/sgbank?sslmode=disable" -verbose up 

migrateup1:
	migrate -path db/migration -database "postgres://root:secret@127.0.0.1:5432/sgbank?sslmode=disable" -verbose up 1

migratedown:
	migrate -path db/migration -database "postgres://root:secret@127.0.0.1:5432/sgbank?sslmode=disable" -verbose down

migratedown1:
	migrate -path db/migration -database "postgres://root:secret@127.0.0.1:5432/sgbank?sslmode=disable" -verbose down 1 

sqlc:
	sqlc generate

test: 
	go test -v -cover ./...

server:
	go run ./cmd/sgbank/main.go

image:
	docker build -t sgbank:latest .

lanch:
	docker run --name sgbank --network sgbank-network -p 8080:8080 -e GIN_MODE=release -e DB_SOURCE="postgresql://root:secret@sgbankpg12:5432/sgbank?sslmode=disable" sgbank:latest

mock:
	mockgen -package mockdb -destination internal/infra/db/mock/store.go github.com/NhutHuyDev/sgbank/internal/infra/db Store

db_docs:
	dbdocs build doc/db.dbml

db_schema:
	dbml2sql --postgres -o doc/schema.sql doc/db.dbml

.PHONY: posgres createdb dropdb migrateup migrateup1 migratedown migratedown1 sqlc test server