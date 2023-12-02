package wsc

import (
	"fmt"

	"github.com/larscom/go-bitvavo/v2/jsond"
	"github.com/larscom/go-bitvavo/v2/log"

	"github.com/goccy/go-json"
	"github.com/larscom/go-bitvavo/v2/util"
	"github.com/smallnest/safemap"
)

type Ticker24hEvent struct {
	// Describes the returned event over the socket.
	Event string `json:"event"`

	// The market which was requested in the subscription.
	Market string `json:"market"`

	// The ticker24h containing the prices etc.
	Ticker24h jsond.Ticker24h `json:"ticker24h"`
}

func (t *Ticker24hEvent) UnmarshalJSON(bytes []byte) error {
	var ticker24hEvent map[string]any

	err := json.Unmarshal(bytes, &ticker24hEvent)
	if err != nil {
		return err
	}

	d := ticker24hEvent["data"].([]any)
	if len(d) != 1 {
		return fmt.Errorf("unexpected length: %d, expected: 1", len(ticker24hEvent))
	}

	var (
		ticker24h = d[0].(map[string]any)

		event          = ticker24hEvent["event"].(string)
		market         = ticker24h["market"].(string)
		open           = ticker24h["open"].(string)
		high           = ticker24h["high"].(string)
		low            = ticker24h["low"].(string)
		last           = ticker24h["last"].(string)
		volume         = ticker24h["volume"].(string)
		volumeQuote    = ticker24h["volumeQuote"].(string)
		bid            = ticker24h["bid"].(string)
		bidSize        = ticker24h["bidSize"].(string)
		ask            = ticker24h["ask"].(string)
		askSize        = ticker24h["askSize"].(string)
		timestamp      = ticker24h["timestamp"].(float64)
		startTimestamp = ticker24h["startTimestamp"].(float64)
		openTimestamp  = ticker24h["openTimestamp"].(float64)
		closeTimestamp = ticker24h["closeTimestamp"].(float64)
	)

	t.Event = event
	t.Market = market
	t.Ticker24h = jsond.Ticker24h{
		Open:           util.IfOrElse(len(open) > 0, func() float64 { return util.MustFloat64(open) }, 0),
		High:           util.IfOrElse(len(high) > 0, func() float64 { return util.MustFloat64(high) }, 0),
		Low:            util.IfOrElse(len(low) > 0, func() float64 { return util.MustFloat64(low) }, 0),
		Last:           util.IfOrElse(len(last) > 0, func() float64 { return util.MustFloat64(last) }, 0),
		Volume:         util.IfOrElse(len(volume) > 0, func() float64 { return util.MustFloat64(volume) }, 0),
		VolumeQuote:    util.IfOrElse(len(volumeQuote) > 0, func() float64 { return util.MustFloat64(volumeQuote) }, 0),
		Bid:            util.IfOrElse(len(bid) > 0, func() float64 { return util.MustFloat64(bid) }, 0),
		BidSize:        util.IfOrElse(len(bidSize) > 0, func() float64 { return util.MustFloat64(bidSize) }, 0),
		Ask:            util.IfOrElse(len(ask) > 0, func() float64 { return util.MustFloat64(ask) }, 0),
		AskSize:        util.IfOrElse(len(askSize) > 0, func() float64 { return util.MustFloat64(askSize) }, 0),
		Timestamp:      int64(timestamp),
		StartTimestamp: int64(startTimestamp),
		OpenTimestamp:  int64(openTimestamp),
		CloseTimestamp: int64(closeTimestamp),
	}

	return nil
}

type ticker24hEventHandler struct {
	writechn chan<- WebSocketMessage
	subs     *safemap.SafeMap[string, chan<- Ticker24hEvent]
}

func newTicker24hEventHandler(writechn chan<- WebSocketMessage) *ticker24hEventHandler {
	return &ticker24hEventHandler{
		writechn: writechn,
		subs:     safemap.New[string, chan<- Ticker24hEvent](),
	}
}

func (t *ticker24hEventHandler) Subscribe(market string, buffSize ...uint64) (<-chan Ticker24hEvent, error) {
	if t.subs.Has(market) {
		return nil, fmt.Errorf("subscription already active for market: %s", market)
	}

	t.writechn <- newWebSocketMessage(actionSubscribe, channelNameTicker24h, market)

	size := util.IfOrElse(len(buffSize) > 0, func() uint64 { return buffSize[0] }, 0)

	chn := make(chan Ticker24hEvent, size)
	t.subs.Set(market, chn)

	return chn, nil
}

func (t *ticker24hEventHandler) Unsubscribe(market string) error {
	sub, exist := t.subs.Get(market)

	if exist {
		t.writechn <- newWebSocketMessage(actionUnsubscribe, channelNameTicker24h, market)
		close(sub)
		t.subs.Remove(market)
		return nil
	}

	return fmt.Errorf("no subscription active for market: %s", market)
}

func (t *ticker24hEventHandler) UnsubscribeAll() error {
	for sub := range t.subs.IterBuffered() {
		market := sub.Key
		if err := t.Unsubscribe(market); err != nil {
			return err
		}
	}
	return nil
}

func (t *ticker24hEventHandler) handleMessage(bytes []byte) {
	var ticker24hEvent *Ticker24hEvent
	if err := json.Unmarshal(bytes, &ticker24hEvent); err != nil {
		log.Logger().Error("Couldn't unmarshal message into Ticker24hEvent", "message", string(bytes))
	} else {
		market := ticker24hEvent.Market
		chn, exist := t.subs.Get(market)
		if exist {
			chn <- *ticker24hEvent
		} else {
			log.Logger().Error("There is no active subscription", "handler", "ticker24h", "market", market)
		}
	}
}

func (t *ticker24hEventHandler) reconnect() {
	for sub := range t.subs.IterBuffered() {
		market := sub.Key
		t.writechn <- newWebSocketMessage(actionSubscribe, channelNameTicker24h, market)
	}
}
