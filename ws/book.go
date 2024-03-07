package ws

import (
	"github.com/google/uuid"
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
	if err := json.Unmarshal(bytes, &bookEvent); err != nil {
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
	subs     *safemap.SafeMap[string, *subscription[BookEvent]]
}

func newBookEventHandler(writechn chan<- WebSocketMessage) *bookEventHandler {
	return &bookEventHandler{
		writechn: writechn,
		subs:     safemap.New[string, *subscription[BookEvent]](),
	}
}

func (b *bookEventHandler) Subscribe(markets []string, buffSize ...uint64) (<-chan BookEvent, error) {
	markets = getUniqueMarkets(markets)

	if err := requireNoSubscription(b.subs, markets); err != nil {
		return nil, err
	}

	var (
		size   = util.IfOrElse(len(buffSize) > 0, func() uint64 { return buffSize[0] }, defaultBuffSize)
		outchn = make(chan BookEvent, size)
		id     = uuid.New()
	)

	for _, market := range markets {
		inchn := make(chan BookEvent, size)
		b.subs.Set(market, newSubscription(id, market, inchn, outchn))
		go relayMessages(inchn, outchn)
	}

	b.writechn <- newWebSocketMessage(actionSubscribe, channelNameBook, markets)

	return outchn, nil
}

func (b *bookEventHandler) Unsubscribe(markets []string) error {
	markets = getUniqueMarkets(markets)

	if err := requireSubscription(b.subs, markets); err != nil {
		return err
	}

	b.writechn <- newWebSocketMessage(actionUnsubscribe, channelNameBook, markets)

	return deleteSubscriptions(b.subs, closeInChannels(b.subs, markets), countSubscriptions(b.subs))
}

func (b *bookEventHandler) UnsubscribeAll() error {
	if err := b.Unsubscribe(b.subs.Keys()); err != nil {
		return err
	}

	return nil
}

func (b *bookEventHandler) handleMessage(bytes []byte) {
	var bookEvent *BookEvent
	if err := json.Unmarshal(bytes, &bookEvent); err != nil {
		log.Err(err).Str("message", string(bytes)).Msg("Couldn't unmarshal message into BookEvent")
	} else {
		market := bookEvent.Market
		sub, exist := b.subs.Get(market)
		if exist {
			sub.inchn <- *bookEvent
		} else {
			log.Error().Str("market", market).Msg("There is no active subscription to handle this BookEvent")
		}
	}
}

func (b *bookEventHandler) reconnect() {
	b.writechn <- newWebSocketMessage(actionSubscribe, channelNameBook, b.subs.Keys())
}
