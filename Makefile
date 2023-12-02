# ws
ticker:
	go run ./examples/ws/ticker/main.go	
candles:
	go run ./examples/ws/candles/main.go
book:
	go run ./examples/ws/book/main.go
ticker24h:
	go run ./examples/ws/ticker24h/main.go		
trades:
	go run ./examples/ws/trades/main.go		
account:
	go run ./examples/ws/account/main.go

# http
time:
	go run ./examples/http/time/main.go	
	
build:
	go build -o ./bin/bitvavo ./bitvavo.go
test:
	go test -v ./.../ --race
