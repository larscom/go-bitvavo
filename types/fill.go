package types

import (
	"github.com/goccy/go-json"
	"github.com/larscom/go-bitvavo/v2/util"
)

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

func (f *Fill) UnmarshalJSON(bytes []byte) error {
	var j map[string]any
	err := json.Unmarshal(bytes, &j)
	if err != nil {
		return err
	}

	var (
		orderId   = j["orderId"].(string)
		fillId    = j["fillId"].(string)
		timestamp = j["timestamp"].(float64)
		amount    = j["amount"].(string)
		side      = j["side"].(string)
		price     = j["price"].(string)
		taker     = j["taker"].(bool)

		// only available if settled is true
		fee         = util.GetOrEmpty[string]("fee", j)
		feeCurrency = util.GetOrEmpty[string]("feeCurrency", j)
	)

	f.OrderId = orderId
	f.FillId = fillId
	f.Timestamp = int64(timestamp)
	f.Amount = util.IfOrElse(len(amount) > 0, func() float64 { return util.MustFloat64(amount) }, 0)
	f.Side = side
	f.Price = util.IfOrElse(len(price) > 0, func() float64 { return util.MustFloat64(price) }, 0)
	f.Taker = taker
	f.Fee = util.IfOrElse(len(fee) > 0, func() float64 { return util.MustFloat64(fee) }, 0)
	f.FeeCurrency = feeCurrency

	return nil
}
