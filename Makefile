build_protoc:
	protoc --go_out=plugins=grpc:. --go_opt=paths=source_relative proto/btchdwallet/wallet.proto

local:
	@echo "> Launch local ..."
	go fmt ./...
#   ENV=local go run ./main.go
	go build -o bin/main ./.
	ENV=local exec ./bin/main

local-test:
	@echo "> Launch local tests ..."
	ENV=test go test ./... -v

print-%:
	@echo '$($*)'

.PHONY: local local-test build_protoc
