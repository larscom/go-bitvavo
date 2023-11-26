candles:
	go run ./examples/candles/main.go
book:
	go run ./examples/book/main.go
ticker:
	go run ./examples/ticker/main.go	
ticker24h:
	go run ./examples/ticker24h/main.go		
trades:
	go run ./examples/trades/main.go		
account:
	go run ./examples/account/main.go
build:
	go build -o ./bin/candles ./examples/candles/main.go
	go build -o ./bin/book ./examples/book/main.go
	go build -o ./bin/ticker ./examples/ticker/main.go
	go build -o ./bin/ticker24h ./examples/ticker24h/main.go
	go build -o ./bin/trades ./examples/trades/main.go
	go build -o ./bin/account ./examples/account/main.go
test:
	go test -v ./.../ --race
