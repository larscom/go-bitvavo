package ws

import (
	"fmt"

	"github.com/larscom/go-bitvavo/v2/types"
	"github.com/rs/zerolog/log"

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
	Ticker24h types.Ticker24h `json:"ticker24h"`
}

func (t *Ticker24hEvent) UnmarshalJSON(bytes []byte) error {
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
		return nil, errSubscriptionAlreadyActive
	}

	t.writechn <- newWebSocketMessage(actionSubscribe, channelNameTicker24h, market)

	size := util.IfOrElse(len(buffSize) > 0, func() uint64 { return buffSize[0] }, defaultBuffSize)

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

	return errNoSubscriptionActive
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
		log.Err(err).Str("message", string(bytes)).Msg("Couldn't unmarshal message into Ticker24hEvent")
	} else {
		market := ticker24hEvent.Market
		chn, exist := t.subs.Get(market)
		if exist {
			chn <- *ticker24hEvent
		} else {
			log.Error().Str("market", market).Msg("There is no active subscription to handle this Ticker24hEvent")
		}
	}
}

func (t *ticker24hEventHandler) reconnect() {
	for sub := range t.subs.IterBuffered() {
		market := sub.Key
		t.writechn <- newWebSocketMessage(actionSubscribe, channelNameTicker24h, market)
	}
}
