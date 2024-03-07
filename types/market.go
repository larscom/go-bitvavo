package types

import (
	"github.com/goccy/go-json"
	"github.com/larscom/go-bitvavo/v2/util"
)

type Market struct {
	// The market itself
	Market string `json:"market"`

	// Enum: "trading" | "halted" | "auction"
	Status string `json:"status"`

	// Base currency, found on the left side of the dash in market.
	Base string `json:"base"`

	// Quote currency, found on the right side of the dash in market.
	Quote string `json:"quote"`

	// Price precision determines how many significant digits are allowed. The rationale behind this is that for higher amounts, smaller price increments are less relevant.
	// Examples of valid prices for precision 5 are: 100010, 11313, 7500.10, 7500.20, 500.12, 0.0012345.
	// Examples of precision 6 are: 11313.1, 7500.11, 7500.25, 500.123, 0.00123456.
	PricePrecision int64 `json:"pricePrecision"`

	// The minimum amount in quote currency (amountQuote or amount * price) for valid orders.
	MinOrderInBaseAsset float64 `json:"minOrderInBaseAsset"`

	// The minimum amount in base currency for valid orders.
	MinOrderInQuoteAsset float64 `json:"minOrderInQuoteAsset"`

	// // The maximum amount in quote currency (amountQuote or amount * price) for valid orders.
	MaxOrderInBaseAsset float64 `json:"maxOrderInBaseAsset"`

	// The maximum amount in base currency for valid orders.
	MaxOrderInQuoteAsset float64 `json:"maxOrderInQuoteAsset"`

	// Allowed order types for this market.
	OrderTypes []string `json:"orderTypes"`
}

func (m *Market) UnmarshalJSON(bytes []byte) error {
	var j map[string]any

	if err := json.Unmarshal(bytes, &j); err != nil {
		return err
	}

	var (
		market               = getOrEmpty[string]("market", j)
		status               = getOrEmpty[string]("status", j)
		base                 = getOrEmpty[string]("base", j)
		quote                = getOrEmpty[string]("quote", j)
		pricePrecision       = getOrEmpty[float64]("pricePrecision", j)
		minOrderInBaseAsset  = getOrEmpty[string]("minOrderInBaseAsset", j)
		minOrderInQuoteAsset = getOrEmpty[string]("minOrderInQuoteAsset", j)
		maxOrderInBaseAsset  = getOrEmpty[string]("maxOrderInBaseAsset", j)
		maxOrderInQuoteAsset = getOrEmpty[string]("maxOrderInQuoteAsset", j)
		orderTypesAny        = getOrEmpty[[]any]("orderTypes", j)
	)

	orderTypes := make([]string, len(orderTypesAny))
	for i := 0; i < len(orderTypesAny); i++ {
		orderTypes[i] = orderTypesAny[i].(string)
	}

	m.Market = market
	m.Status = status
	m.Base = base
	m.Quote = quote
	m.PricePrecision = int64(pricePrecision)
	m.MinOrderInBaseAsset = util.IfOrElse(len(minOrderInBaseAsset) > 0, func() float64 { return util.MustFloat64(minOrderInBaseAsset) }, 0)
	m.MinOrderInQuoteAsset = util.IfOrElse(len(minOrderInQuoteAsset) > 0, func() float64 { return util.MustFloat64(minOrderInQuoteAsset) }, 0)
	m.MaxOrderInBaseAsset = util.IfOrElse(len(maxOrderInBaseAsset) > 0, func() float64 { return util.MustFloat64(maxOrderInBaseAsset) }, 0)
	m.MaxOrderInQuoteAsset = util.IfOrElse(len(maxOrderInQuoteAsset) > 0, func() float64 { return util.MustFloat64(maxOrderInQuoteAsset) }, 0)
	m.OrderTypes = orderTypes

	return nil
}
