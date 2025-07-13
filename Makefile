proto-gen:
	@echo "Generating Go protobuf files..."
	cd src/proto && protoc --go_out=. --go-grpc_out=. \
		--go-grpc_opt=paths=source_relative --go_opt=paths=source_relative *.proto

server-run:
	@echo "Building server..."
	go run src/server/server.go

client-run:
	@echo "Building client..."
	go run src/client/client.go

tidy:
	go mod tidy

.PHONY: proto-gen server-run client-run tidy
