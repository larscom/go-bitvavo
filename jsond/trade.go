package jsond

type Trade struct {
	// The trade ID of the returned trade (UUID).
	Id string `json:"id"`

	// The amount in base currency for which the trade has been made.
	Amount float64 `json:"amount"`

	// The price in quote currency for which the trade has been made.
	Price float64 `json:"price"`

	// The side for the taker.
	// Enum: "buy" | "sell"
	Side string `json:"side"`

	// Timestamp in unix milliseconds.
	Timestamp int64 `json:"timestamp"`
}
