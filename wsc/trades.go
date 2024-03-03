package wsc

import (
	"log/slog"

	"github.com/larscom/go-bitvavo/v2/types"

	"github.com/goccy/go-json"
	"github.com/larscom/go-bitvavo/v2/util"
	"github.com/smallnest/safemap"
)

type TradesEvent struct {
	// Describes the returned event over the socket.
	Event string `json:"event"`

	// The market which was requested in the subscription.
	Market string `json:"market"`

	// The trade containing the price, side etc.
	Trade types.Trade `json:"trade"`
}

func (t *TradesEvent) UnmarshalJSON(bytes []byte) error {
	if err := t.Trade.UnmarshalJSON(bytes); err != nil {
		return err
	}

	var tradesEvent map[string]any
	err := json.Unmarshal(bytes, &tradesEvent)
	if err != nil {
		return err
	}

	var (
		event  = tradesEvent["event"].(string)
		market = tradesEvent["market"].(string)
	)

	t.Event = event
	t.Market = market

	return nil
}

type tradesEventHandler struct {
	writechn chan<- WebSocketMessage
	subs     *safemap.SafeMap[string, chan<- TradesEvent]
}

func newTradesEventHandler(writechn chan<- WebSocketMessage) *tradesEventHandler {
	return &tradesEventHandler{
		writechn: writechn,
		subs:     safemap.New[string, chan<- TradesEvent](),
	}
}

func (t *tradesEventHandler) Subscribe(market string, buffSize ...uint64) (<-chan TradesEvent, error) {
	if t.subs.Has(market) {
		return nil, ErrSubscriptionAlreadyActive
	}

	t.writechn <- newWebSocketMessage(actionSubscribe, channelNameTrades, market)

	size := util.IfOrElse(len(buffSize) > 0, func() uint64 { return buffSize[0] }, DefaultBuffSize)

	chn := make(chan TradesEvent, size)
	t.subs.Set(market, chn)

	return chn, nil
}

func (t *tradesEventHandler) Unsubscribe(market string) error {
	sub, exist := t.subs.Get(market)

	if exist {
		t.writechn <- newWebSocketMessage(actionUnsubscribe, channelNameTrades, market)
		close(sub)
		t.subs.Remove(market)
		return nil
	}

	return ErrNoSubscriptionActive
}

func (t *tradesEventHandler) UnsubscribeAll() error {
	for sub := range t.subs.IterBuffered() {
		market := sub.Key
		if err := t.Unsubscribe(market); err != nil {
			return err
		}
	}
	return nil
}

func (t *tradesEventHandler) handleMessage(bytes []byte) {
	var tradeEvent *TradesEvent
	if err := json.Unmarshal(bytes, &tradeEvent); err != nil {
		slog.Error("Couldn't unmarshal message into TradesEvent", "message", string(bytes))
	} else {
		market := tradeEvent.Market
		chn, exist := t.subs.Get(market)
		if exist {
			chn <- *tradeEvent
		} else {
			slog.Error("There is no active subscription", "handler", "trades", "market", market)
		}
	}
}

func (t *tradesEventHandler) reconnect() {
	for sub := range t.subs.IterBuffered() {
		market := sub.Key
		t.writechn <- newWebSocketMessage(actionSubscribe, channelNameTrades, market)
	}
}
