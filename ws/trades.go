package ws

import (
	"github.com/google/uuid"
	"github.com/larscom/go-bitvavo/v2/types"
	"github.com/rs/zerolog/log"

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
	if err := json.Unmarshal(bytes, &tradesEvent); err != nil {
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
	subs     *safemap.SafeMap[string, *subscription[TradesEvent]]
}

func newTradesEventHandler(writechn chan<- WebSocketMessage) *tradesEventHandler {
	return &tradesEventHandler{
		writechn: writechn,
		subs:     safemap.New[string, *subscription[TradesEvent]](),
	}
}

func (t *tradesEventHandler) Subscribe(markets []string, buffSize ...uint64) (<-chan TradesEvent, error) {
	markets = getUniqueMarkets(markets)

	if err := requireNoSubscription(t.subs, markets); err != nil {
		return nil, err
	}

	var (
		size   = util.IfOrElse(len(buffSize) > 0, func() uint64 { return buffSize[0] }, defaultBuffSize)
		outchn = make(chan TradesEvent, int(size)*len(markets))
		id     = uuid.New()
	)

	for _, market := range markets {
		inchn := make(chan TradesEvent, size)
		t.subs.Set(market, newSubscription(id, market, inchn, outchn))
		go relayMessages(inchn, outchn)
	}

	t.writechn <- newWebSocketMessage(actionSubscribe, channelNameTrades, markets)

	return outchn, nil
}

func (t *tradesEventHandler) Unsubscribe(markets []string) error {
	markets = getUniqueMarkets(markets)

	if err := requireSubscription(t.subs, markets); err != nil {
		return err
	}

	t.writechn <- newWebSocketMessage(actionUnsubscribe, channelNameTrades, markets)

	return deleteSubscriptions(t.subs, markets)
}

func (t *tradesEventHandler) UnsubscribeAll() error {
	if err := t.Unsubscribe(t.subs.Keys()); err != nil {
		return err
	}

	return nil
}

func (t *tradesEventHandler) handleMessage(bytes []byte) {
	var tradeEvent *TradesEvent
	if err := json.Unmarshal(bytes, &tradeEvent); err != nil {
		log.Err(err).Str("message", string(bytes)).Msg("Couldn't unmarshal message into TradesEvent")
	} else {
		market := tradeEvent.Market
		sub, exist := t.subs.Get(market)
		if exist {
			sub.inchn <- *tradeEvent
		} else {
			log.Debug().Str("market", market).Msg("There is no active subscription to handle this TradesEvent")
		}
	}
}

func (t *tradesEventHandler) reconnect() {
	t.writechn <- newWebSocketMessage(actionSubscribe, channelNameTrades, t.subs.Keys())
}
