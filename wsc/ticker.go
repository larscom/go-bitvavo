package wsc

import (
	"fmt"

	"github.com/larscom/go-bitvavo/v2/jsond"
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
	Ticker jsond.Ticker `json:"ticker"`
}

func (t *TickerEvent) UnmarshalJSON(bytes []byte) error {
	var tickerEvent map[string]string

	err := json.Unmarshal(bytes, &tickerEvent)
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
	t.Ticker = jsond.Ticker{
		BestBid:     util.IfOrElse(len(bestBid) > 0, func() float64 { return util.MustFloat64(bestBid) }, 0),
		BestBidSize: util.IfOrElse(len(bestBidSize) > 0, func() float64 { return util.MustFloat64(bestBidSize) }, 0),
		BestAsk:     util.IfOrElse(len(bestAsk) > 0, func() float64 { return util.MustFloat64(bestAsk) }, 0),
		BestAskSize: util.IfOrElse(len(bestAskSize) > 0, func() float64 { return util.MustFloat64(bestAskSize) }, 0),
		LastPrice:   util.IfOrElse(len(lastPrice) > 0, func() float64 { return util.MustFloat64(lastPrice) }, 0),
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

func (t *tickerEventHandler) Subscribe(market string, buffSize ...uint64) (<-chan TickerEvent, error) {
	if t.subs.Has(market) {
		return nil, fmt.Errorf("subscription already active for market: %s", market)
	}

	t.writechn <- newWebSocketMessage(actionSubscribe, channelNameTicker, market)

	size := util.IfOrElse(len(buffSize) > 0, func() uint64 { return buffSize[0] }, 0)

	chn := make(chan TickerEvent, size)
	t.subs.Set(market, chn)

	return chn, nil
}

func (t *tickerEventHandler) Unsubscribe(market string) error {
	sub, exist := t.subs.Get(market)

	if exist {
		t.writechn <- newWebSocketMessage(actionUnsubscribe, channelNameTicker, market)
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
		t.writechn <- newWebSocketMessage(actionSubscribe, channelNameTicker, market)
	}
}