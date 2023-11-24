package bitvavo

import (
	"fmt"

	"github.com/larscom/go-bitvavo/v2/log"

	"github.com/goccy/go-json"
	"github.com/larscom/go-bitvavo/v2/util"
	"github.com/smallnest/safemap"
)

type TradesEvent struct {
	// Describes the returned event over the socket
	Event string `json:"event"`
	// The market which was requested in the subscription
	Market string `json:"market"`
	// The trade containing the price, side etc
	Trade Trade `json:"trade"`
}

type Trade struct {
	// The trade ID of the returned trade (UUID)
	Id string `json:"id"`
	// The amount in base currency for which the trade has been made
	Amount float64 `json:"amount"`
	// The price in quote currency for which the trade has been made
	Price float64 `json:"price"`
	// The side for the taker
	// Enum: "buy" | "sell"
	Side string `json:"side"`
	// Timestamp in unix milliseconds
	Timestamp int64 `json:"timestamp"`
}

func (t *TradesEvent) UnmarshalJSON(bytes []byte) error {
	var tradesEvent map[string]any
	err := json.Unmarshal(bytes, &tradesEvent)
	if err != nil {
		return err
	}

	var (
		event     = tradesEvent["event"].(string)
		market    = tradesEvent["market"].(string)
		id        = tradesEvent["id"].(string)
		amount    = tradesEvent["amount"].(string)
		price     = tradesEvent["price"].(string)
		side      = tradesEvent["side"].(string)
		timestamp = tradesEvent["timestamp"].(float64)
	)

	t.Event = event
	t.Market = market
	t.Trade = Trade{
		Id:        id,
		Amount:    util.IfOrElse(len(amount) > 0, func() float64 { return util.MustFloat64(amount) }, ZERO),
		Price:     util.IfOrElse(len(price) > 0, func() float64 { return util.MustFloat64(price) }, ZERO),
		Side:      side,
		Timestamp: int64(timestamp),
	}

	return nil
}

type tradesWsHandler struct {
	writechn chan<- WebSocketMessage
	subs     *safemap.SafeMap[string, chan<- TradesEvent]
}

func newTradesWsHandler(writechn chan<- WebSocketMessage) *tradesWsHandler {
	return &tradesWsHandler{
		writechn: writechn,
		subs:     safemap.New[string, chan<- TradesEvent](),
	}
}

func (t *tradesWsHandler) Subscribe(market string) (<-chan TradesEvent, error) {
	if t.subs.Has(market) {
		return nil, fmt.Errorf("subscription already active for market: %s", market)
	}

	t.writechn <- newWebSocketMessage(ActionSubscribe, ChannelNameTrades, market)

	chn := make(chan TradesEvent)
	t.subs.Set(market, chn)

	return chn, nil
}

func (t *tradesWsHandler) Unsubscribe(market string) error {
	sub, exist := t.subs.Get(market)

	if exist {
		t.writechn <- newWebSocketMessage(ActionUnsubscribe, ChannelNameTrades, market)
		close(sub)
		t.subs.Remove(market)
		return nil
	}

	return fmt.Errorf("no subscription active for market: %s", market)
}

func (t *tradesWsHandler) UnsubscribeAll() error {
	for sub := range t.subs.IterBuffered() {
		market := sub.Key
		if err := t.Unsubscribe(market); err != nil {
			return err
		}
	}
	return nil
}

func (t *tradesWsHandler) handleMessage(bytes []byte) {
	var tradeEvent *TradesEvent
	if err := json.Unmarshal(bytes, &tradeEvent); err != nil {
		log.Logger().Error("Couldn't unmarshal message into TradesEvent", "message", string(bytes))
	} else {
		market := tradeEvent.Market
		chn, exist := t.subs.Get(market)
		if exist {
			chn <- *tradeEvent
		} else {
			log.Logger().Error("There is no active subscription", "handler", "trades", "market", market)
		}
	}
}

func (t *tradesWsHandler) reconnect() {
	for sub := range t.subs.IterBuffered() {
		market := sub.Key
		t.writechn <- newWebSocketMessage(ActionSubscribe, ChannelNameTrades, market)
	}
}
