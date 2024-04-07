package ws

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/larscom/go-bitvavo/v2/types"
	csmap "github.com/mhmtszr/concurrent-swiss-map"
	"github.com/rs/zerolog/log"

	"github.com/goccy/go-json"
	"github.com/larscom/go-bitvavo/v2/util"
)

type Ticker24hEvent struct {
	// Describes the returned event over the socket.
	Event string `json:"event"`

	// The market which was requested in the subscription.
	Market string `json:"market"`

	// The ticker24h containing the prices etc.
	Ticker24h types.Ticker24h `json:"ticker24h"`
}

func (t *Ticker24hEvent) UnmarshalJSON(bytes []byte) error {
	var ticker24hEvent map[string]any

	if err := json.Unmarshal(bytes, &ticker24hEvent); err != nil {
		return err
	}

	data := ticker24hEvent["data"].([]any)
	if len(data) != 1 {
		return fmt.Errorf("unexpected length: %d, expected: 1", len(ticker24hEvent))
	}

	var (
		ticker24h = data[0].(map[string]any)
		event     = ticker24hEvent["event"].(string)
		market    = ticker24h["market"].(string)
	)

	ticker24hBytes, err := json.Marshal(ticker24h)
	if err != nil {
		return err
	}

	if err := t.Ticker24h.UnmarshalJSON(ticker24hBytes); err != nil {
		return err
	}

	t.Event = event
	t.Market = market

	return nil
}

type ticker24hEventHandler struct {
	writechn chan<- WebSocketMessage
	subs     *csmap.CsMap[string, *subscription[Ticker24hEvent]]
}

func newTicker24hEventHandler(writechn chan<- WebSocketMessage) *ticker24hEventHandler {
	return &ticker24hEventHandler{
		writechn: writechn,
		subs:     csmap.Create[string, *subscription[Ticker24hEvent]](),
	}
}

func (t *ticker24hEventHandler) Subscribe(markets []string, buffSize ...uint64) (<-chan Ticker24hEvent, error) {
	markets = getUniqueMarkets(markets)

	if err := requireNoSubscription(t.subs, markets); err != nil {
		return nil, err
	}
	var (
		size   = util.IfOrElse(len(buffSize) > 0, func() uint64 { return buffSize[0] }, defaultBuffSize)
		outchn = make(chan Ticker24hEvent, int(size)*len(markets))
		id     = uuid.New()
	)

	for _, market := range markets {
		inchn := make(chan Ticker24hEvent, size)
		t.subs.Store(market, newSubscription(id, market, inchn, outchn))
		go relayMessages(inchn, outchn)
	}

	t.writechn <- newWebSocketMessage(actionSubscribe, channelNameTicker24h, markets)

	return outchn, nil
}

func (t *ticker24hEventHandler) Unsubscribe(markets []string) error {
	markets = getUniqueMarkets(markets)

	if err := requireSubscription(t.subs, markets); err != nil {
		return err
	}

	t.writechn <- newWebSocketMessage(actionUnsubscribe, channelNameTicker24h, markets)

	return deleteSubscriptions(t.subs, markets)
}

func (t *ticker24hEventHandler) UnsubscribeAll() error {
	if err := t.Unsubscribe(getSubscriptionKeys(t.subs)); err != nil {
		return err
	}

	return nil
}

func (t *ticker24hEventHandler) handleMessage(_ WsEvent, bytes []byte) {
	var ticker24hEvent *Ticker24hEvent
	if err := json.Unmarshal(bytes, &ticker24hEvent); err != nil {
		log.Err(err).Str("message", string(bytes)).Msg("Couldn't unmarshal message into Ticker24hEvent")
	} else {
		market := ticker24hEvent.Market
		sub, exist := t.subs.Load(market)
		if exist {
			sub.inchn <- *ticker24hEvent
		} else {
			log.Debug().Str("market", market).Msg("There is no active subscription to handle this Ticker24hEvent")
		}
	}
}

func (t *ticker24hEventHandler) reconnect() {
	t.writechn <- newWebSocketMessage(actionSubscribe, channelNameTicker24h, getSubscriptionKeys(t.subs))
}
