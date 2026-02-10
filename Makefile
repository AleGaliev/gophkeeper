.PHONY: test clean

format-code:
	@echo "🚀 Formt code..."
	go fmt ./...

gen-proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --go_opt=default_api_level=API_OPAQUE internal/pkg/proto/gophkeeper.proto

go-tidy:
	go mod tidy

build-agent:
	@echo "🚀 Building agent..."
	go build -o cmd/agent/agent cmd/agent/*.go

build-server:
	@echo "🚀 Building server..."
	go build -o cmd/server/server cmd/server/*.go

server-migration-up:
	@echo "🚀 Up migration..."
	cmd/server/server migrate -db "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" -up

server-migration-down:
	@echo "🚀 Down migration..."
	cmd/server/server migrate -db "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" -down

server-start:
	@echo "🚀 Start Server..."
	cmd/server/server start -db "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" 
