package types

type Ticker24h struct {
	// The open price of the 24 hour period.
	Open float64 `json:"open"`

	// The highest price for which a trade occurred in the 24 hour period.
	High float64 `json:"high"`

	// The lowest price for which a trade occurred in the 24 hour period.
	Low float64 `json:"low"`

	// The last price for which a trade occurred in the 24 hour period.
	Last float64 `json:"last"`

	// The total volume of the 24 hour period in base currency.
	Volume float64 `json:"volume"`

	// The total volume of the 24 hour period in quote currency.
	VolumeQuote float64 `json:"volumeQuote"`

	// The best (highest) bid offer at the current moment.
	Bid float64 `json:"bid"`

	// The size of the best (highest) bid offer.
	BidSize float64 `json:"bidSize"`

	// The best (lowest) ask offer at the current moment.
	Ask float64 `json:"ask"`

	// The size of the best (lowest) ask offer.
	AskSize float64 `json:"askSize"`

	// Timestamp in unix milliseconds.
	Timestamp int64 `json:"timestamp"`

	// Start timestamp in unix milliseconds.
	StartTimestamp int64 `json:"startTimestamp"`

	// Open timestamp in unix milliseconds.
	OpenTimestamp int64 `json:"openTimestamp"`

	// Close timestamp in unix milliseconds.
	CloseTimestamp int64 `json:"closeTimestamp"`
}
