package ws

import (
	"fmt"
	"strings"

	"github.com/larscom/go-bitvavo/v2/types"
	"github.com/larscom/go-bitvavo/v2/util"
	"github.com/rs/zerolog/log"

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
	Candle types.Candle `json:"candle"`
}

func (c *CandlesEvent) UnmarshalJSON(bytes []byte) error {
	var candleEvent map[string]any
	err := json.Unmarshal(bytes, &candleEvent)
	if err != nil {
		return err
	}

	var (
		event    = candleEvent["event"].(string)
		market   = candleEvent["market"].(string)
		interval = candleEvent["interval"].(string)
		candle   = candleEvent["candle"].([]any)
	)

	if len(candle) != 1 {
		return fmt.Errorf("unexpected length: %d, expected: 1", len(candle))
	}

	candleBytes, err := json.Marshal(candle[0])
	if err != nil {
		return err
	}

	if err := c.Candle.UnmarshalJSON(candleBytes); err != nil {
		return err
	}

	c.Event = event
	c.Market = market
	c.Interval = interval

	return nil
}

type CandlesEventHandler interface {
	// Subscribe to market with interval.
	// You can set the buffSize for this channel.
	//
	// If you have many subscriptions at once you may need to increase the buffSize
	//
	// Default buffSize: 50
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

	key := createKey(market, interval)
	if c.subs.Has(key) {
		return nil, errSubscriptionAlreadyActive
	}

	c.writechn <- newCandleWebSocketMessage(actionSubscribe, market, interval)

	size := util.IfOrElse(len(buffSize) > 0, func() uint64 { return buffSize[0] }, defaultBuffSize)

	chn := make(chan CandlesEvent, size)
	c.subs.Set(key, chn)

	return chn, nil
}

func (c *candlesEventHandler) Unsubscribe(market string, interval string) error {
	key := createKey(market, interval)
	sub, exist := c.subs.Get(key)

	if exist {
		c.writechn <- newCandleWebSocketMessage(actionUnsubscribe, market, interval)
		close(sub)
		c.subs.Remove(key)
		return nil
	}

	return errNoSubscriptionActive
}

func (c *candlesEventHandler) UnsubscribeAll() error {
	for sub := range c.subs.IterBuffered() {
		market, interval := parseKey(sub.Key)
		if err := c.Unsubscribe(market, interval); err != nil {
			return err
		}
	}
	return nil
}

func (c *candlesEventHandler) handleMessage(bytes []byte) {
	var candleEvent *CandlesEvent
	if err := json.Unmarshal(bytes, &candleEvent); err != nil {
		log.Err(err).Str("message", string(bytes)).Msg("Couldn't unmarshal message into CandlesEvent")
	} else {
		var (
			market   = candleEvent.Market
			interval = candleEvent.Interval
			key      = createKey(market, interval)
		)

		chn, exist := c.subs.Get(key)
		if exist {
			chn <- *candleEvent
		} else {
			log.Error().Str("market", market).Msg("There is no active subscription to handle this CandlesEvent")
		}
	}
}

func (c *candlesEventHandler) reconnect() {
	for sub := range c.subs.IterBuffered() {
		market, interval := parseKey(sub.Key)
		c.writechn <- newCandleWebSocketMessage(actionSubscribe, market, interval)
	}
}

func parseKey(key string) (string, string) {
	parts := strings.Split(key, "_")
	market := parts[0]
	interval := parts[1]
	return market, interval
}

func createKey(market string, interval string) string {
	return fmt.Sprintf("%s_%s", market, interval)
}
