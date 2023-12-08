package types

import (
	"fmt"
	"net/url"
	"time"

	"github.com/goccy/go-json"
	"github.com/larscom/go-bitvavo/v2/util"
)

type TradeParams struct {
	// Return the limit most recent trades only.
	// Default: 500
	Limit uint64 `json:"limit"`

	// Return limit trades executed after start.
	Start time.Time `json:"start"`

	// Return limit trades executed before end.
	End time.Time `json:"end"`

	// Return limit trades executed after tradeIdFrom was made.
	TradeIdFrom string `json:"tradeIdFrom"`

	// Return limit trades executed before tradeIdTo was made.
	TradeIdTo string `json:"tradeIdTo"`
}

func (t *TradeParams) ToParams() url.Values {
	params := make(url.Values)
	if t.Limit > 0 {
		params.Add("limit", fmt.Sprint(t.Limit))
	}
	if !t.Start.IsZero() {
		params.Add("start", fmt.Sprint(t.Start.UnixMilli()))
	}
	if !t.End.IsZero() {
		params.Add("end", fmt.Sprint(t.End.UnixMilli()))
	}
	if t.TradeIdFrom != "" {
		params.Add("tradeIdFrom", t.TradeIdFrom)
	}
	if t.TradeIdTo != "" {
		params.Add("tradeIdTo", t.TradeIdTo)
	}
	return params
}

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

func (t *Trade) UnmarshalJSON(bytes []byte) error {
	var j map[string]any

	err := json.Unmarshal(bytes, &j)
	if err != nil {
		return err
	}

	var (
		id        = j["id"].(string)
		amount    = j["amount"].(string)
		price     = j["price"].(string)
		side      = j["side"].(string)
		timestamp = j["timestamp"].(float64)
	)

	t.Id = id
	t.Amount = util.IfOrElse(len(amount) > 0, func() float64 { return util.MustFloat64(amount) }, 0)
	t.Price = util.IfOrElse(len(price) > 0, func() float64 { return util.MustFloat64(price) }, 0)
	t.Side = side
	t.Timestamp = int64(timestamp)

	return nil
}
