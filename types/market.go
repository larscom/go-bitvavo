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

	err := json.Unmarshal(bytes, &j)
	if err != nil {
		return err
	}

	var (
		market               = j["market"].(string)
		status               = j["status"].(string)
		base                 = j["base"].(string)
		quote                = j["quote"].(string)
		pricePrecision       = j["pricePrecision"].(float64)
		minOrderInBaseAsset  = j["minOrderInBaseAsset"].(string)
		minOrderInQuoteAsset = j["minOrderInQuoteAsset"].(string)
		maxOrderInBaseAsset  = j["maxOrderInBaseAsset"].(string)
		maxOrderInQuoteAsset = j["maxOrderInQuoteAsset"].(string)
		orderTypesAny        = j["orderTypes"].([]any)
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
	m.MinOrderInBaseAsset = util.MustFloat64(minOrderInBaseAsset)
	m.MinOrderInQuoteAsset = util.MustFloat64(minOrderInQuoteAsset)
	m.MaxOrderInBaseAsset = util.MustFloat64(maxOrderInBaseAsset)
	m.MaxOrderInQuoteAsset = util.MustFloat64(maxOrderInQuoteAsset)
	m.OrderTypes = orderTypes

	return nil
}
