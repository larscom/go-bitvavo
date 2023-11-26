run:
	go run ./examples/candles/main.go
build:
	go build -o ./bin/candles ./examples/candles/main.go
	go build -o ./bin/account ./examples/account/main.go
test:
	go test -v ./.../ --race
