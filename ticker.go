package bitvavo

import (
	"fmt"

	"github.com/larscom/go-bitvavo/v2/log"

	"github.com/goccy/go-json"
	"github.com/larscom/go-bitvavo/v2/util"
	"github.com/smallnest/safemap"
)

type TickerEvent struct {
	// Describes the returned event over the socket.
	Event string `json:"event"`

	// The market which was requested in the subscription.
	Market string `json:"market"`

	// The ticker containing the prices.
	Ticker Ticker `json:"ticker"`
}

type Ticker struct {
	// The price of the best (highest) bid offer available, only sent when either bestBid or bestBidSize has changed.
	BestBid float64 `json:"bestBid"`

	// The size of the best (highest) bid offer available, only sent when either bestBid or bestBidSize has changed.
	BestBidSize float64 `json:"bestBidSize"`

	// The price of the best (lowest) ask offer available, only sent when either bestAsk or bestAskSize has changed.
	BestAsk float64 `json:"bestAsk"`

	// The size of the best (lowest) ask offer available, only sent when either bestAsk or bestAskSize has changed.
	BestAskSize float64 `json:"bestAskSize"`

	// The last price for which a trade has occurred, only sent when lastPrice has changed.
	LastPrice float64 `json:"lastPrice"`
}

func (t *TickerEvent) UnmarshalJSON(data []byte) error {
	var tickerEvent map[string]string

	err := json.Unmarshal(data, &tickerEvent)
	if err != nil {
		return err
	}

	var (
		market      = tickerEvent["market"]
		bestBid     = tickerEvent["bestBid"]
		bestBidSize = tickerEvent["bestBidSize"]
		bestAsk     = tickerEvent["bestAsk"]
		bestAskSize = tickerEvent["bestAskSize"]
		lastPrice   = tickerEvent["lastPrice"]
	)

	t.Market = market
	t.Ticker = Ticker{
		BestBid:     util.IfOrElse(len(bestBid) > 0, func() float64 { return util.MustFloat64(bestBid) }, ZERO),
		BestBidSize: util.IfOrElse(len(bestBidSize) > 0, func() float64 { return util.MustFloat64(bestBidSize) }, ZERO),
		BestAsk:     util.IfOrElse(len(bestAsk) > 0, func() float64 { return util.MustFloat64(bestAsk) }, ZERO),
		BestAskSize: util.IfOrElse(len(bestAskSize) > 0, func() float64 { return util.MustFloat64(bestAskSize) }, ZERO),
		LastPrice:   util.IfOrElse(len(lastPrice) > 0, func() float64 { return util.MustFloat64(lastPrice) }, ZERO),
	}

	return nil
}

type tickerEventHandler struct {
	writechn chan<- WebSocketMessage
	subs     *safemap.SafeMap[string, chan<- TickerEvent]
}

func newTickerEventHandler(writechn chan<- WebSocketMessage) *tickerEventHandler {
	return &tickerEventHandler{
		writechn: writechn,
		subs:     safemap.New[string, chan<- TickerEvent](),
	}
}

func (t *tickerEventHandler) Subscribe(market string, buffSize uint64) (<-chan TickerEvent, error) {
	if t.subs.Has(market) {
		return nil, fmt.Errorf("subscription already active for market: %s", market)
	}

	t.writechn <- newWebSocketMessage(ActionSubscribe, ChannelNameTicker, market)

	chn := make(chan TickerEvent, buffSize)
	t.subs.Set(market, chn)

	return chn, nil
}

func (t *tickerEventHandler) Unsubscribe(market string) error {
	sub, exist := t.subs.Get(market)

	if exist {
		t.writechn <- newWebSocketMessage(ActionUnsubscribe, ChannelNameTicker, market)
		close(sub)
		t.subs.Remove(market)
		return nil
	}

	return fmt.Errorf("no subscription active for market: %s", market)
}

func (t *tickerEventHandler) UnsubscribeAll() error {
	for sub := range t.subs.IterBuffered() {
		market := sub.Key
		if err := t.Unsubscribe(market); err != nil {
			return err
		}
	}
	return nil
}

func (t *tickerEventHandler) handleMessage(bytes []byte) {
	var tickerEvent *TickerEvent
	if err := json.Unmarshal(bytes, &tickerEvent); err != nil {
		log.Logger().Error("Couldn't unmarshal message into TickerEvent", "message", string(bytes))
	} else {
		market := tickerEvent.Market
		chn, exist := t.subs.Get(market)
		if exist {
			chn <- *tickerEvent
		} else {
			log.Logger().Error("There is no active subscription", "handler", "ticker", "market", market)
		}
	}
}

func (t *tickerEventHandler) reconnect() {
	for sub := range t.subs.IterBuffered() {
		market := sub.Key
		t.writechn <- newWebSocketMessage(ActionSubscribe, ChannelNameTicker, market)
	}
}
