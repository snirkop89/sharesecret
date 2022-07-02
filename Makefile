.PHONY: run
run:
	DATA_FILE_PATH="./secret.json" go run ./cmd/api/

.PHONY: build
build:
	go build -o bin/secret-app ./cmd/api/