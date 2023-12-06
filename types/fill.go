package types

type Fill struct {
	// The id of the order on which has been filled
	OrderId string `json:"orderId"`

	// The id of the returned fill
	FillId string `json:"fillId"`

	// The current timestamp in milliseconds since 1 Jan 1970
	Timestamp int64 `json:"timestamp"`

	// The amount in base currency for which the trade has been made
	Amount float64 `json:"amount"`

	// The side for the taker
	// Enum: "buy" | "sell"
	Side string `json:"side"`

	// The price in quote currency for which the trade has been made
	Price float64 `json:"price"`

	// True for takers, false for makers
	Taker bool `json:"taker"`

	// The amount of fee that has been paid. Value is negative for rebates. Only available if settled is true
	Fee float64 `json:"fee"`

	// Currency in which the fee has been paid. Only available if settled is true
	FeeCurrency string `json:"feeCurrency"`
}
