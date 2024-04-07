package ws

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/larscom/go-bitvavo/v2/types"
	"github.com/larscom/go-bitvavo/v2/util"
	csmap "github.com/mhmtszr/concurrent-swiss-map"
	"github.com/rs/zerolog/log"

	"github.com/goccy/go-json"
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
	if err := json.Unmarshal(bytes, &candleEvent); err != nil {
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
	// Subscribe to markets with interval.
	// You can set the buffSize for this channel.
	//
	// If you have many subscriptions at once you may need to increase the buffSize
	//
	// Default buffSize: 50
	Subscribe(markets []string, interval string, buffSize ...uint64) (<-chan CandlesEvent, error)

	// Unsubscribe from markets with interval
	Unsubscribe(markets []string, interval string) error

	// Unsubscribe from every market with interval
	UnsubscribeAll() error
}

type candlesEventHandler struct {
	writechn chan<- WebSocketMessage
	subs     *csmap.CsMap[string, *subscription[CandlesEvent]]
}

func newCandlesEventHandler(writechn chan<- WebSocketMessage) *candlesEventHandler {
	return &candlesEventHandler{
		writechn: writechn,
		subs:     csmap.Create[string, *subscription[CandlesEvent]](),
	}
}

func newCandleWebSocketMessage(action Action, markets []string, interval string) WebSocketMessage {
	return WebSocketMessage{
		Action: action.Value,
		Channels: []Channel{
			{
				Name:      channelNameCandles.Value,
				Markets:   markets,
				Intervals: []string{interval},
			},
		},
	}
}

func (c *candlesEventHandler) Subscribe(markets []string, interval string, buffSize ...uint64) (<-chan CandlesEvent, error) {
	markets = getUniqueMarkets(markets)
	keys := c.createKeys(markets, interval)

	for i, key := range keys {
		if c.subs.Has(key) {
			return nil, errSubscriptionAlreadyActive(markets[i])
		}
	}

	var (
		size   = util.IfOrElse(len(buffSize) > 0, func() uint64 { return buffSize[0] }, defaultBuffSize)
		outchn = make(chan CandlesEvent, int(size)*len(keys))
		id     = uuid.New()
	)

	for i, key := range keys {
		inchn := make(chan CandlesEvent, size)
		c.subs.Store(key, newSubscription(id, markets[i], inchn, outchn))
		go relayMessages(inchn, outchn)
	}

	c.writechn <- newCandleWebSocketMessage(actionSubscribe, markets, interval)

	return outchn, nil
}

func (c *candlesEventHandler) Unsubscribe(markets []string, interval string) error {
	markets = getUniqueMarkets(markets)

	keys := c.createKeys(markets, interval)

	for i, key := range keys {
		if !c.subs.Has(key) {
			return errNoSubscriptionActive(markets[i])
		}
	}

	c.writechn <- newCandleWebSocketMessage(actionUnsubscribe, markets, interval)

	return deleteSubscriptions(c.subs, keys)
}

func (c *candlesEventHandler) UnsubscribeAll() error {
	for interval, markets := range c.getIntervalMarkets() {
		if err := c.Unsubscribe(markets, interval); err != nil {
			return err
		}
	}

	return nil
}

func (c *candlesEventHandler) handleMessage(_ WsEvent, bytes []byte) {
	var candleEvent *CandlesEvent
	if err := json.Unmarshal(bytes, &candleEvent); err != nil {
		log.Err(err).Str("message", string(bytes)).Msg("Couldn't unmarshal message into CandlesEvent")
	} else {
		var (
			market   = candleEvent.Market
			interval = candleEvent.Interval
			key      = c.createKey(market, interval)
		)

		sub, exist := c.subs.Load(key)
		if exist {
			sub.inchn <- *candleEvent
		} else {
			log.Debug().Str("market", market).Msg("There is no active subscription to handle this CandlesEvent")
		}
	}
}

func (c *candlesEventHandler) reconnect() {
	for interval, markets := range c.getIntervalMarkets() {
		c.writechn <- newCandleWebSocketMessage(actionSubscribe, markets, interval)
	}
}

func (c *candlesEventHandler) getIntervalMarkets() map[string][]string {
	m := make(map[string][]string)

	c.subs.Range(func(key string, _ *subscription[CandlesEvent]) (stop bool) {
		market, interval := c.parseKey(key)
		m[interval] = append(m[interval], market)
		return false
	})

	return m
}

func (c *candlesEventHandler) parseKey(key string) (string, string) {
	parts := strings.Split(key, "_")
	market := parts[0]
	interval := parts[1]
	return market, interval
}

func (c *candlesEventHandler) createKey(market string, interval string) string {
	return fmt.Sprintf("%s_%s", market, interval)
}

func (c *candlesEventHandler) createKeys(markets []string, interval string) []string {
	keys := make([]string, len(markets))
	for i := 0; i < len(keys); i++ {
		keys[i] = c.createKey(markets[i], interval)
	}
	return keys
}
