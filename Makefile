.PHONY: test clean

format-code:
	@echo "🚀 Formt code..."
	go fmt ./...

gen-proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --go_opt=default_api_level=API_OPAQUE internal/pkg/proto/gophkeeper.proto

go-tidy:
	go mod tidy

build-client:
	@echo "🚀 Building agent..."
	go build -o cmd/client/client cmd/client/*.go

create-secret:
	@echo "🚀 Create Secret..."
	cmd/client/client -u alex -p alex -a localhost:8081 create -name "testi" -secret-type "text" -text "lkdmfkdnmfk" -description "kjdnfkjndfkjndf"

update-secret:
	@echo "🚀 Update Secret..."
	cmd/client/client -u alex -p alex -a localhost:8081 update -name "testi" -secret-type "text" -description "dfdfdfdfdf" -text "dfdfdf"


delete-secret:
	@echo "🚀 Update Secret..."
	cmd/client/client -u alex -p alex -a localhost:8081 delete -name "testi" -secret-type "text"


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
	cmd/server/server -log-level "debug" start -db "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"

build-and-server-start: build-server server-start
