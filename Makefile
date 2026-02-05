.PHONY: test clean

format-code:
	@echo "🚀 Formt code..."
	go fmt ./...

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