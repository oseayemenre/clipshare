build:
	@go build -o bin/main cmd/api/*.go

run: build
	@bin/main
