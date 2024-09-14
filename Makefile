build:
	@go build -o bin/es

test:
	@go test ./...

run: build
	@./bin/es
