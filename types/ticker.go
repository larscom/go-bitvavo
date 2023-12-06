package types

type Ticker struct {
	// The price of the best (highest) bid offer available, only sent when either bestBid or bestBidSize has changed.
	BestBid float64 `json:"bestBid"`

	// The size of the best (highest) bid offer available, only sent when either bestBid or bestBidSize has changed.
	BestBidSize float64 `json:"bestBidSize"`

	// The price of the best (lowest) ask offer available, only sent when either bestAsk or bestAskSize has changed.
	BestAsk float64 `json:"bestAsk"`

	// The size of the best (lowest) ask offer available, only sent when either bestAsk or bestAskSize has changed.
	BestAskSize float64 `json:"bestAskSize"`

	// The last price for which a trade has occurred, only sent when lastPrice has changed.
	LastPrice float64 `json:"lastPrice"`
}
