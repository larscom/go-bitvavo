package types

import (
	"github.com/goccy/go-json"
	"github.com/larscom/go-bitvavo/v2/util"
)

type Fill struct {
	// The id of the returned fill
	FillId string `json:"fillId"`

	// The id of the order on which has been filled
	OrderId string `json:"orderId"`

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

	// True when the fee has been deducted and the bought/sold currency is available for further trading.
	// Fills are settled almost instantly.
	Settled bool `json:"settled"`
}

func (f *Fill) UnmarshalJSON(bytes []byte) error {
	var j map[string]any

	if err := json.Unmarshal(bytes, &j); err != nil {
		return err
	}

	var (
		fillId = getOrEmpty[string]("fillId", j)
		id     = getOrEmpty[string]("id", j)

		orderId     = getOrEmpty[string]("orderId", j)
		timestamp   = getOrEmpty[float64]("timestamp", j)
		amount      = getOrEmpty[string]("amount", j)
		side        = getOrEmpty[string]("side", j)
		price       = getOrEmpty[string]("price", j)
		taker       = getOrEmpty[bool]("taker", j)
		fee         = getOrEmpty[string]("fee", j)
		feeCurrency = getOrEmpty[string]("feeCurrency", j)
		settled     = getOrEmpty[bool]("settled", j)
	)

	f.OrderId = orderId
	f.FillId = util.IfOrElse(len(fillId) > 0, func() string { return fillId }, id)
	f.Timestamp = int64(timestamp)
	f.Amount = util.IfOrElse(len(amount) > 0, func() float64 { return util.MustFloat64(amount) }, 0)
	f.Side = side
	f.Price = util.IfOrElse(len(price) > 0, func() float64 { return util.MustFloat64(price) }, 0)
	f.Taker = taker
	f.Fee = util.IfOrElse(len(fee) > 0, func() float64 { return util.MustFloat64(fee) }, 0)
	f.FeeCurrency = feeCurrency
	f.Settled = settled

	return nil
}
