build:
	@go build -o bin/es

test:
	@go test ./... -v

run: build
	@./bin/es
