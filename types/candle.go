package types

import (
	"fmt"
	"net/url"
	"time"

	"github.com/goccy/go-json"
	"github.com/larscom/go-bitvavo/v2/util"
)

type CandleParams struct {
	// Return the limit most recent candlesticks only.
	// Default: 1440
	Limit uint64 `json:"limit"`

	// Return limit candlesticks for trades made after start.
	Start time.Time `json:"start"`

	// Return limit candlesticks for trades made before end.
	End time.Time `json:"end"`
}

func (c *CandleParams) Params() url.Values {
	params := make(url.Values)
	if c.Limit > 0 {
		params.Add("limit", fmt.Sprint(c.Limit))
	}
	if !c.Start.IsZero() {
		params.Add("start", fmt.Sprint(c.Start.UnixMilli()))
	}
	if !c.End.IsZero() {
		params.Add("end", fmt.Sprint(c.End.UnixMilli()))
	}
	return params
}

type Candle struct {
	// Timestamp in unix milliseconds.
	Timestamp int64   `json:"timestamp"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    float64 `json:"volume"`
}

func (c *Candle) UnmarshalJSON(bytes []byte) error {
	var j []any
	if err := json.Unmarshal(bytes, &j); err != nil {
		return err
	}

	c.Timestamp = int64(j[0].(float64))
	c.Open = util.MustFloat64(j[1].(string))
	c.High = util.MustFloat64(j[2].(string))
	c.Low = util.MustFloat64(j[3].(string))
	c.Close = util.MustFloat64(j[4].(string))
	c.Volume = util.MustFloat64(j[5].(string))

	return nil
}
