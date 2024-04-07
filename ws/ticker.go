package ws

import (
	"github.com/google/uuid"
	"github.com/larscom/go-bitvavo/v2/types"
	csmap "github.com/mhmtszr/concurrent-swiss-map"
	"github.com/rs/zerolog/log"

	"github.com/goccy/go-json"
	"github.com/larscom/go-bitvavo/v2/util"
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
	if err := json.Unmarshal(bytes, &tickerEvent); err != nil {
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
	subs     *csmap.CsMap[string, *subscription[TickerEvent]]
}

func newTickerEventHandler(writechn chan<- WebSocketMessage) *tickerEventHandler {
	return &tickerEventHandler{
		writechn: writechn,
		subs:     csmap.Create[string, *subscription[TickerEvent]](),
	}
}

func (t *tickerEventHandler) Subscribe(markets []string, buffSize ...uint64) (<-chan TickerEvent, error) {
	markets = getUniqueMarkets(markets)

	if err := requireNoSubscription(t.subs, markets); err != nil {
		return nil, err
	}

	var (
		size   = util.IfOrElse(len(buffSize) > 0, func() uint64 { return buffSize[0] }, defaultBuffSize)
		outchn = make(chan TickerEvent, int(size)*len(markets))
		id     = uuid.New()
	)

	for _, market := range markets {
		inchn := make(chan TickerEvent, size)
		t.subs.Store(market, newSubscription(id, market, inchn, outchn))
		go relayMessages(inchn, outchn)
	}

	t.writechn <- newWebSocketMessage(actionSubscribe, channelNameTicker, markets)

	return outchn, nil
}

func (t *tickerEventHandler) Unsubscribe(markets []string) error {
	markets = getUniqueMarkets(markets)

	if err := requireSubscription(t.subs, markets); err != nil {
		return err
	}

	t.writechn <- newWebSocketMessage(actionUnsubscribe, channelNameTicker, markets)

	return deleteSubscriptions(t.subs, markets)
}

func (t *tickerEventHandler) UnsubscribeAll() error {
	if err := t.Unsubscribe(getSubscriptionKeys(t.subs)); err != nil {
		return err
	}

	return nil
}

func (t *tickerEventHandler) handleMessage(_ WsEvent, bytes []byte) {
	var tickerEvent *TickerEvent
	if err := json.Unmarshal(bytes, &tickerEvent); err != nil {
		log.Err(err).Str("message", string(bytes)).Msg("Couldn't unmarshal message into TickerEvent")
	} else {
		market := tickerEvent.Market
		sub, exist := t.subs.Load(market)
		if exist {
			sub.inchn <- *tickerEvent
		} else {
			log.Debug().Str("market", market).Msg("There is no active subscription to handle this TickerEvent")
		}
	}
}

func (t *tickerEventHandler) reconnect() {
	t.writechn <- newWebSocketMessage(actionSubscribe, channelNameTicker, getSubscriptionKeys(t.subs))
}
