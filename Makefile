build:
	go build -o ./build/main ./cmd/main.go

run: 
	npx nodemon --exec go run ./cmd/main.go --signal SIGTERM 

.PHONY: build run