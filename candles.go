package bitvavo

import (
	"fmt"
	"strings"

	"github.com/larscom/go-bitvavo/v2/log"

	"github.com/goccy/go-json"
	"github.com/larscom/go-bitvavo/v2/util"
	"github.com/smallnest/safemap"
)

type CandlesEvent struct {
	// Describes the returned event over the socket
	Event string `json:"event"`
	// The market which was requested in the subscription
	Market string `json:"market"`
	//The interval which was requested in the subscription
	Interval string `json:"interval"`
	// The candle in the defined time period
	Candle Candle `json:"candle"`
}

type Candle struct {
	// Timestamp in unix milliseconds
	Timestamp int64   `json:"timestamp"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    float64 `json:"volume"`
}

func (c *Candle) UnmarshalJSON(data []byte) error {
	var event [][]any

	err := json.Unmarshal(data, &event)
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

type CandlesWsHandler interface {
	// Subscribe to market with interval
	Subscribe(market string, interval string) (<-chan CandlesEvent, error)

	// Unsubscribe from market with interval
	Unsubscribe(market string, interval string) error

	// Unsubscribe from every market
	UnsubscribeAll() error
}

type candleWsHandler struct {
	writechn chan<- WebSocketMessage
	subs     *safemap.SafeMap[string, chan<- CandlesEvent]
}

func newCandleWsHandler(writechn chan<- WebSocketMessage) *candleWsHandler {
	return &candleWsHandler{
		writechn: writechn,
		subs:     safemap.New[string, chan<- CandlesEvent](),
	}
}

func newCandleWebSocketMessage(action Action, market string, interval string) WebSocketMessage {
	return WebSocketMessage{
		Action: action.Value,
		Channels: []Channel{
			{
				Name:      ChannelNameCandles.Value,
				Markets:   []string{market},
				Intervals: []string{interval},
			},
		},
	}
}

func (c *candleWsHandler) Subscribe(market string, interval string) (<-chan CandlesEvent, error) {

	key := getMapKey(market, interval)
	if c.subs.Has(key) {
		return nil, fmt.Errorf("subscription already active for market: %s with interval: %s", market, interval)
	}

	c.writechn <- newCandleWebSocketMessage(ActionSubscribe, market, interval)

	chn := make(chan CandlesEvent)
	c.subs.Set(key, chn)

	return chn, nil
}

func (c *candleWsHandler) Unsubscribe(market string, interval string) error {
	key := getMapKey(market, interval)
	sub, exist := c.subs.Get(key)

	if exist {
		c.writechn <- newCandleWebSocketMessage(ActionUnsubscribe, market, interval)
		close(sub)
		c.subs.Remove(key)
		return nil
	}

	return fmt.Errorf("no subscription active for market: %s with interval: %s", market, interval)
}

func (c *candleWsHandler) UnsubscribeAll() error {
	for sub := range c.subs.IterBuffered() {
		market, interval := getMapKeyValue(sub.Key)
		if err := c.Unsubscribe(market, interval); err != nil {
			return err
		}
	}
	return nil
}

func (c *candleWsHandler) handleMessage(bytes []byte) {
	var candleEvent *CandlesEvent
	if err := json.Unmarshal(bytes, &candleEvent); err != nil {
		log.Logger().Error("Couldn't unmarshal message into CandlesEvent", "message", string(bytes))
	} else {
		var (
			market   = candleEvent.Market
			interval = candleEvent.Interval
			key      = getMapKey(market, interval)
		)

		chn, exist := c.subs.Get(key)
		if exist {
			chn <- *candleEvent
		} else {
			log.Logger().Error("There is no active subscription", "handler", "candles", "market", market, "interval", interval)
		}
	}
}

func (c *candleWsHandler) reconnect() {
	for sub := range c.subs.IterBuffered() {
		market, interval := getMapKeyValue(sub.Key)
		c.writechn <- newCandleWebSocketMessage(ActionSubscribe, market, interval)
	}
}

func getMapKeyValue(key string) (string, string) {
	parts := strings.Split(key, "_")
	market := parts[0]
	interval := parts[1]
	return market, interval
}

func getMapKey(market string, interval string) string {
	return fmt.Sprintf("%s_%s", market, interval)
}
