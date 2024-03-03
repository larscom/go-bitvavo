package ws

import (
	"log/slog"

	"github.com/larscom/go-bitvavo/v2/types"

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
	Ticker types.Ticker `json:"ticker"`
}

func (t *TickerEvent) UnmarshalJSON(bytes []byte) error {
	if err := t.Ticker.UnmarshalJSON(bytes); err != nil {
		return err
	}

	var tickerEvent map[string]string
	err := json.Unmarshal(bytes, &tickerEvent)
	if err != nil {
		return err
	}

	var (
		market = tickerEvent["market"]
		event  = tickerEvent["event"]
	)

	t.Event = event
	t.Market = market

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
		return nil, errSubscriptionAlreadyActive
	}

	t.writechn <- newWebSocketMessage(actionSubscribe, channelNameTicker, market)

	size := util.IfOrElse(len(buffSize) > 0, func() uint64 { return buffSize[0] }, defaultBuffSize)

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

	return errNoSubscriptionActive
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
		slog.Error("Couldn't unmarshal message into TickerEvent", "message", string(bytes))
	} else {
		market := tickerEvent.Market
		chn, exist := t.subs.Get(market)
		if exist {
			chn <- *tickerEvent
		} else {
			slog.Error("There is no active subscription", "handler", "ticker", "market", market)
		}
	}
}

func (t *tickerEventHandler) reconnect() {
	for sub := range t.subs.IterBuffered() {
		market := sub.Key
		t.writechn <- newWebSocketMessage(actionSubscribe, channelNameTicker, market)
	}
}
