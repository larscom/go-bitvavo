run:
	go run ./examples/main.go
build:
	go build -o ./bin/example ./examples/main.go
test:
	go test -v ./.../ --race
