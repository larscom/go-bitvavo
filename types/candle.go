package types

import (
	"fmt"

	"github.com/goccy/go-json"
	"github.com/larscom/go-bitvavo/v2/util"
)

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
	var event [][]any

	err := json.Unmarshal(bytes, &event)
	if err != nil {
		return err
	}
	if len(event) != 1 {
		return fmt.Errorf("unexpected length: %d, expected: 1", len(event))
	}

	candle := event[0]

	c.Timestamp = int64(candle[0].(float64))
	c.Open = util.MustFloat64(candle[1].(string))
	c.High = util.MustFloat64(candle[2].(string))
	c.Low = util.MustFloat64(candle[3].(string))
	c.Close = util.MustFloat64(candle[4].(string))
	c.Volume = util.MustFloat64(candle[5].(string))

	return nil
}
