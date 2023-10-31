
bin/server: server/server.go
	@echo "Building server.go..."
	@go build -o bin/server server/server.go

bin/client: client/client.go
	@echo "Building client.go..."
	@go build -o bin/client client/client.go

build: bin/server bin/client

runServer: bin/server
	@echo "Running server..."
	@./bin/server

runClient: bin/client
	@echo "Running client..."
	@./bin/client

clean:
	@echo "Deleting bin directory..."
	@rm -rf bin

genProtoc: proto/ChittyChat.proto
	@echo "Generating proto files..."
	@protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/ChittyChat.proto

cleanProtoc:
	@echo "Deleting proto files..."
	@rm -f proto/ChittyChat.pb.go proto/ChittyChat_grpc.pb.go

.PHONY: build runServer runClient clean genProtoc cleanProtoc

