.PHONY: help test server docker up down logs logs-chat logs-db evans proto lint scylla client migrate test-db test-db/down test-db/run
test: 
	go test -v ./...

server:
	go run cmd/chat/main.go

client:
	go run cmd/client/main.go --roomID=C828351E-ED3F-4D1B-AE05-293F92D95B36
	
docker:
	docker build -t chat .

scylla:
	docker run --name scylla --hostname scylladb -d -p 9042:9042 scylladb/scylla --smp 1

test-db:
	docker compose -f internal/db/compose.yaml up -d --wait
	bash -c "trap '$(MAKE) test-db/down' EXIT; $(MAKE) test-db/run"

test-db/down:
	docker compose -f internal/db/compose.yaml down -v

test-db/run:
	go test -tags testdb -v --coverprofile=coverage.out -coverpkg ./... ./internal/db/keyspace/...

migrate:
	go run cmd/migrate/main.go

up:
	docker-compose up -d

down:
	docker-compose down -v

logs:
	docker-compose logs -f

logs-chat:
	docker-compose logs -f chat

logs-db:
	docker-compose logs -f cassandra

evans:
	evans -r
	
proto:
	rm -rf pkg/proto/gen/*.go
	protoc --proto_path=pkg/proto --go_out=pkg/proto/gen --go_opt=paths=source_relative \
    --go-grpc_out=pkg/proto/gen --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=pkg/proto/gen --grpc-gateway_opt=paths=source_relative \
     pkg/proto/*.proto

lint:
	golangci-lint run --fix
