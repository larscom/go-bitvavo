package ws

import (
	"github.com/larscom/go-bitvavo/v2/types"
	"github.com/rs/zerolog/log"

	"github.com/goccy/go-json"
	"github.com/larscom/go-bitvavo/v2/util"
	"github.com/smallnest/safemap"
)

type BookEvent struct {
	// Describes the returned event over the socket.
	Event string `json:"event"`

	// The market which was requested in the subscription.
	Market string `json:"market"`

	// The book containing the bids and asks.
	Book types.Book `json:"book"`
}

func (b *BookEvent) UnmarshalJSON(bytes []byte) error {
	if err := b.Book.UnmarshalJSON(bytes); err != nil {
		return err
	}

	var bookEvent map[string]any
	err := json.Unmarshal(bytes, &bookEvent)
	if err != nil {
		return err
	}

	var (
		event  = bookEvent["event"].(string)
		market = bookEvent["market"].(string)
	)

	b.Event = event
	b.Market = market

	return nil
}

type bookEventHandler struct {
	writechn chan<- WebSocketMessage
	subs     *safemap.SafeMap[string, chan<- BookEvent]
}

func newBookEventHandler(writechn chan<- WebSocketMessage) *bookEventHandler {
	return &bookEventHandler{
		writechn: writechn,
		subs:     safemap.New[string, chan<- BookEvent](),
	}
}

func (t *bookEventHandler) Subscribe(market string, buffSize ...uint64) (<-chan BookEvent, error) {
	if t.subs.Has(market) {
		return nil, errSubscriptionAlreadyActive
	}

	t.writechn <- newWebSocketMessage(actionSubscribe, channelNameBook, market)

	size := util.IfOrElse(len(buffSize) > 0, func() uint64 { return buffSize[0] }, defaultBuffSize)

	chn := make(chan BookEvent, size)
	t.subs.Set(market, chn)

	return chn, nil
}

func (t *bookEventHandler) Unsubscribe(market string) error {
	sub, exist := t.subs.Get(market)

	if exist {
		t.writechn <- newWebSocketMessage(actionUnsubscribe, channelNameBook, market)
		close(sub)
		t.subs.Remove(market)
		return nil
	}

	return errNoSubscriptionActive
}

func (t *bookEventHandler) UnsubscribeAll() error {
	for sub := range t.subs.IterBuffered() {
		market := sub.Key
		if err := t.Unsubscribe(market); err != nil {
			return err
		}
	}
	return nil
}

func (t *bookEventHandler) handleMessage(bytes []byte) {
	var bookEvent *BookEvent
	if err := json.Unmarshal(bytes, &bookEvent); err != nil {
		log.Err(err).Str("message", string(bytes)).Msg("Couldn't unmarshal message into BookEvent")
	} else {
		market := bookEvent.Market
		chn, exist := t.subs.Get(market)
		if exist {
			chn <- *bookEvent
		} else {
			log.Error().Str("market", market).Msg("There is no active subscription to handle this BookEvent")
		}
	}
}

func (t *bookEventHandler) reconnect() {
	for sub := range t.subs.IterBuffered() {
		market := sub.Key
		t.writechn <- newWebSocketMessage(actionSubscribe, channelNameBook, market)
	}
}
