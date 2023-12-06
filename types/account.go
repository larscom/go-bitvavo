package types

import (
	"github.com/goccy/go-json"
	"github.com/larscom/go-bitvavo/v2/util"
)

type Account struct {
	Fees Fee `json:"fees"`
}

type Fee struct {
	// Fee for trades that take liquidity from the order book.
	Taker float64 `json:"taker"`

	// Fee for trades that add liquidity to the order book.
	Maker float64 `json:"maker"`

	// Your trading volume in the last 30 days measured in EUR.
	Volume float64 `json:"volume"`
}

func (f *Fee) UnmarshalJSON(bytes []byte) error {
	var j map[string]string

	err := json.Unmarshal(bytes, &j)
	if err != nil {
		return err
	}

	var (
		taker  = j["taker"]
		maker  = j["maker"]
		volume = j["volume"]
	)

	f.Taker = util.IfOrElse(len(taker) > 0, func() float64 { return util.MustFloat64(taker) }, 0)
	f.Maker = util.IfOrElse(len(maker) > 0, func() float64 { return util.MustFloat64(maker) }, 0)
	f.Volume = util.IfOrElse(len(volume) > 0, func() float64 { return util.MustFloat64(volume) }, 0)

	return nil
}
