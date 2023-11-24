package bitvavo

import (
	"fmt"

	"github.com/larscom/go-bitvavo/v2/log"

	"github.com/goccy/go-json"
	"github.com/larscom/go-bitvavo/v2/util"
	"github.com/smallnest/safemap"
)

type Ticker24hEvent struct {
	// Describes the returned event over the socket
	Event string `json:"event"`
	// The market which was requested in the subscription
	Market string `json:"market"`
	// The ticker24h containing the prices etc
	Ticker24h Ticker24h `json:"ticker24h"`
}

type Ticker24h struct {
	// The open price of the 24 hour period
	Open float64 `json:"open"`
	// The highest price for which a trade occurred in the 24 hour period
	High float64 `json:"high"`
	// The lowest price for which a trade occurred in the 24 hour period
	Low float64 `json:"low"`
	// The last price for which a trade occurred in the 24 hour period
	Last float64 `json:"last"`
	// The total volume of the 24 hour period in base currency
	Volume float64 `json:"volume"`
	// The total volume of the 24 hour period in quote currency
	VolumeQuote float64 `json:"volumeQuote"`
	// The best (highest) bid offer at the current moment
	Bid float64 `json:"bid"`
	// The size of the best (highest) bid offer
	BidSize float64 `json:"bidSize"`
	// The best (lowest) ask offer at the current moment
	Ask float64 `json:"ask"`
	// The size of the best (lowest) ask offer
	AskSize float64 `json:"askSize"`
	// Timestamp in unix milliseconds
	Timestamp int64 `json:"timestamp"`
	// Start timestamp in unix milliseconds
	StartTimestamp int64 `json:"startTimestamp"`
	// Open timestamp in unix milliseconds
	OpenTimestamp int64 `json:"openTimestamp"`
	// Close timestamp in unix milliseconds
	CloseTimestamp int64 `json:"closeTimestamp"`
}

func (t *Ticker24hEvent) UnmarshalJSON(bytes []byte) error {
	// {"event":"ticker24h","ticker24hEvent":[{"market":"ETH-EUR","startTimestamp":1700425282396,"timestamp":1700511682396,"open":"1815.1","openTimestamp":1700425292390,"high":"1890.2","low":"1813.1","last":"1863.8","closeTimestamp":1700511637320,"bid":"1862.7","bidSize":"1.719","ask":"1864.3","askSize":"7.9779","volume":"3629.1566404","volumeQuote":"6720833.300920673"}]}
	var ticker24hEvent map[string]any
	err := json.Unmarshal(bytes, &ticker24hEvent)
	if err != nil {
		return err
	}
	data := ticker24hEvent["data"].([]any)
	if len(data) != 1 {
		return fmt.Errorf("unexpected length: %d, expected: 1", len(ticker24hEvent))
	}
	var (
		ticker24h = data[0].(map[string]any)

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
	t.Ticker24h = Ticker24h{
		Open:           util.IfOrElse(len(open) > 0, func() float64 { return util.MustFloat64(open) }, ZERO),
		High:           util.IfOrElse(len(high) > 0, func() float64 { return util.MustFloat64(high) }, ZERO),
		Low:            util.IfOrElse(len(low) > 0, func() float64 { return util.MustFloat64(low) }, ZERO),
		Last:           util.IfOrElse(len(last) > 0, func() float64 { return util.MustFloat64(last) }, ZERO),
		Volume:         util.IfOrElse(len(volume) > 0, func() float64 { return util.MustFloat64(volume) }, ZERO),
		VolumeQuote:    util.IfOrElse(len(volumeQuote) > 0, func() float64 { return util.MustFloat64(volumeQuote) }, ZERO),
		Bid:            util.IfOrElse(len(bid) > 0, func() float64 { return util.MustFloat64(bid) }, ZERO),
		BidSize:        util.IfOrElse(len(bidSize) > 0, func() float64 { return util.MustFloat64(bidSize) }, ZERO),
		Ask:            util.IfOrElse(len(ask) > 0, func() float64 { return util.MustFloat64(ask) }, ZERO),
		AskSize:        util.IfOrElse(len(askSize) > 0, func() float64 { return util.MustFloat64(askSize) }, ZERO),
		Timestamp:      int64(timestamp),
		StartTimestamp: int64(startTimestamp),
		OpenTimestamp:  int64(openTimestamp),
		CloseTimestamp: int64(closeTimestamp),
	}

	return nil
}

type ticker24hWsHandler struct {
	writechn chan<- WebSocketMessage
	subs     *safemap.SafeMap[string, chan<- Ticker24hEvent]
}

func newTicker24hWsHandler(writechn chan<- WebSocketMessage) *ticker24hWsHandler {
	return &ticker24hWsHandler{
		writechn: writechn,
		subs:     safemap.New[string, chan<- Ticker24hEvent](),
	}
}

func (t *ticker24hWsHandler) Subscribe(market string) (<-chan Ticker24hEvent, error) {
	if t.subs.Has(market) {
		return nil, fmt.Errorf("subscription already active for market: %s", market)
	}

	t.writechn <- newWebSocketMessage(ActionSubscribe, ChannelNameTicker24h, market)

	chn := make(chan Ticker24hEvent)
	t.subs.Set(market, chn)

	return chn, nil
}

func (t *ticker24hWsHandler) Unsubscribe(market string) error {
	sub, exist := t.subs.Get(market)

	if exist {
		t.writechn <- newWebSocketMessage(ActionUnsubscribe, ChannelNameTicker24h, market)
		close(sub)
		t.subs.Remove(market)
		return nil
	}

	return fmt.Errorf("no subscription active for market: %s", market)
}

func (t *ticker24hWsHandler) UnsubscribeAll() error {
	for sub := range t.subs.IterBuffered() {
		market := sub.Key
		if err := t.Unsubscribe(market); err != nil {
			return err
		}
	}
	return nil
}

func (t *ticker24hWsHandler) handleMessage(bytes []byte) {
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

func (t *ticker24hWsHandler) reconnect() {
	for sub := range t.subs.IterBuffered() {
		market := sub.Key
		t.writechn <- newWebSocketMessage(ActionSubscribe, ChannelNameTicker24h, market)
	}
}
