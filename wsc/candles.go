package wsc

import (
	"fmt"
	"strings"

	"github.com/larscom/go-bitvavo/v2/jsond"
	"github.com/larscom/go-bitvavo/v2/log"
	"github.com/larscom/go-bitvavo/v2/util"

	"github.com/goccy/go-json"
	"github.com/smallnest/safemap"
)

type CandlesEvent struct {
	// Describes the returned event over the socket.
	Event string `json:"event"`

	// The market which was requested in the subscription.
	Market string `json:"market"`

	// The interval which was requested in the subscription.
	Interval string `json:"interval"`

	// The candle in the defined time period.
	Candle jsond.Candle `json:"candle"`
}

type CandlesEventHandler interface {
	// Subscribe to market with interval.
	// You can set the buffSize for this channel.
	Subscribe(market string, interval string, buffSize ...uint64) (<-chan CandlesEvent, error)

	// Unsubscribe from market with interval
	Unsubscribe(market string, interval string) error

	// Unsubscribe from every market
	UnsubscribeAll() error
}

type candlesEventHandler struct {
	writechn chan<- WebSocketMessage
	subs     *safemap.SafeMap[string, chan<- CandlesEvent]
}

func newCandlesEventHandler(writechn chan<- WebSocketMessage) *candlesEventHandler {
	return &candlesEventHandler{
		writechn: writechn,
		subs:     safemap.New[string, chan<- CandlesEvent](),
	}
}

func newCandleWebSocketMessage(action Action, market string, interval string) WebSocketMessage {
	return WebSocketMessage{
		Action: action.Value,
		Channels: []Channel{
			{
				Name:      channelNameCandles.Value,
				Markets:   []string{market},
				Intervals: []string{interval},
			},
		},
	}
}

func (c *candlesEventHandler) Subscribe(market string, interval string, buffSize ...uint64) (<-chan CandlesEvent, error) {

	key := getMapKey(market, interval)
	if c.subs.Has(key) {
		return nil, fmt.Errorf("subscription already active for market: %s with interval: %s", market, interval)
	}

	c.writechn <- newCandleWebSocketMessage(actionSubscribe, market, interval)

	size := util.IfOrElse(len(buffSize) > 0, func() uint64 { return buffSize[0] }, 0)

	chn := make(chan CandlesEvent, size)
	c.subs.Set(key, chn)

	return chn, nil
}

func (c *candlesEventHandler) Unsubscribe(market string, interval string) error {
	key := getMapKey(market, interval)
	sub, exist := c.subs.Get(key)

	if exist {
		c.writechn <- newCandleWebSocketMessage(actionUnsubscribe, market, interval)
		close(sub)
		c.subs.Remove(key)
		return nil
	}

	return fmt.Errorf("no subscription active for market: %s with interval: %s", market, interval)
}

func (c *candlesEventHandler) UnsubscribeAll() error {
	for sub := range c.subs.IterBuffered() {
		market, interval := getMapKeyValue(sub.Key)
		if err := c.Unsubscribe(market, interval); err != nil {
			return err
		}
	}
	return nil
}

func (c *candlesEventHandler) handleMessage(bytes []byte) {
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

func (c *candlesEventHandler) reconnect() {
	for sub := range c.subs.IterBuffered() {
		market, interval := getMapKeyValue(sub.Key)
		c.writechn <- newCandleWebSocketMessage(actionSubscribe, market, interval)
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
