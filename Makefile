test: 
	go test -v ./...

run:
	go run ./cmd/app/main.go

docker:
	docker build -t chat .

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

test/integration:
	docker compose -f test/integration/docker-compose.yaml up -d --wait
	bash -c "trap '$(MAKE) test/integration/down' EXIT; $(MAKE) test/integration/run"

test/integration/down:
	docker compose -f test/integration/docker-compose.yaml down -v

test/integration/run:
	POSTGRES_PASSWORD=integration \
	USER_SERVICE_SYMMETRIC_KEY=12345678901234567890123456789012 \
	go test -tags integration -v --coverprofile=coverage.out -coverpkg ./... ./test/integration

lint:
	golangci-lint run --fix
