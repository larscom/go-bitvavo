package ws

import (
	"time"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/larscom/go-bitvavo/v2/crypto"
	"github.com/larscom/go-bitvavo/v2/util"
	"github.com/rs/zerolog/log"

	"github.com/larscom/go-bitvavo/v2/types"
	"github.com/smallnest/safemap"
)

type OrderEvent struct {
	// Describes the returned event over the socket.
	Event string `json:"event"`

	// The market which was requested in the subscription.
	Market string `json:"market"`

	// The order itself.
	Order types.Order `json:"order"`
}

func (o *OrderEvent) UnmarshalJSON(bytes []byte) error {
	if err := o.Order.UnmarshalJSON(bytes); err != nil {
		return err
	}

	var orderEvent map[string]any
	if err := json.Unmarshal(bytes, &orderEvent); err != nil {
		return err
	}

	var (
		market = orderEvent["market"].(string)
		event  = orderEvent["event"].(string)
	)

	o.Market = market
	o.Event = event

	return nil
}

type FillEvent struct {
	// Describes the returned event over the socket
	Event string `json:"event"`
	// The market which was requested in the subscription
	Market string `json:"market"`
	// The fill itself
	Fill types.Fill `json:"fill"`
}

func (f *FillEvent) UnmarshalJSON(bytes []byte) error {
	if err := f.Fill.UnmarshalJSON(bytes); err != nil {
		return err
	}

	var fillEvent map[string]any
	if err := json.Unmarshal(bytes, &fillEvent); err != nil {
		return err

	}

	var (
		market = fillEvent["market"].(string)
		event  = fillEvent["event"].(string)
	)

	f.Market = market
	f.Event = event

	return nil
}

type AccountEventHandler interface {
	// Subscribe to markets.
	// You can set the buffSize for the channel.
	// If you have many subscriptions at once you may need to increase the buffSize
	// Default buffSize: 50
	Subscribe(markets []string, buffSize ...uint64) (<-chan OrderEvent, <-chan FillEvent, error)

	// Unsubscribe from markets.
	Unsubscribe(markets []string) error

	// Unsubscribe from every market.
	UnsubscribeAll() error
}

type accountSubscription struct {
	id     uuid.UUID
	market string

	orderinchn  chan<- OrderEvent
	orderoutchn chan OrderEvent

	fillinchn  chan<- FillEvent
	filloutchn chan FillEvent
}

func newAccountSubscription(
	id uuid.UUID,
	market string,
	orderinchn chan<- OrderEvent,
	orderoutchn chan OrderEvent,
	fillinchn chan<- FillEvent,
	filloutchn chan FillEvent,
) *accountSubscription {
	return &accountSubscription{
		id:          id,
		market:      market,
		orderinchn:  orderinchn,
		orderoutchn: orderoutchn,
		fillinchn:   fillinchn,
		filloutchn:  filloutchn,
	}
}

type accountEventHandler struct {
	apiKey        string
	apiSecret     string
	authenticated bool
	authchn       chan bool
	writechn      chan<- WebSocketMessage
	subs          *safemap.SafeMap[string, *accountSubscription]
}

func newAccountEventHandler(apiKey string, apiSecret string, writechn chan<- WebSocketMessage) *accountEventHandler {
	return &accountEventHandler{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		writechn:  writechn,
		authchn:   make(chan bool),
		subs:      safemap.New[string, *accountSubscription](),
	}
}

func (a *accountEventHandler) Subscribe(markets []string, buffSize ...uint64) (<-chan OrderEvent, <-chan FillEvent, error) {
	markets = getUniqueMarkets(markets)

	if err := requireNoSubscription(a.subs, markets); err != nil {
		return nil, nil, err
	}

	if err := a.withAuth(func() {
		a.writechn <- newWebSocketMessage(actionSubscribe, channelNameAccount, markets)
	}); err != nil {
		return nil, nil, err
	}

	var (
		size        = util.IfOrElse(len(buffSize) > 0, func() uint64 { return buffSize[0] }, defaultBuffSize)
		orderoutchn = make(chan OrderEvent, size)
		filloutchn  = make(chan FillEvent, size)
		id          = uuid.New()
	)

	for _, market := range markets {
		orderinchn := make(chan OrderEvent, size)
		fillinchn := make(chan FillEvent, size)

		a.subs.Set(market, newAccountSubscription(id, market, orderinchn, orderoutchn, fillinchn, filloutchn))

		go relayMessages(orderinchn, orderoutchn)
		go relayMessages(fillinchn, filloutchn)
	}

	return orderoutchn, filloutchn, nil

}

func (a *accountEventHandler) Unsubscribe(markets []string) error {
	markets = getUniqueMarkets(markets)

	if err := requireSubscription(a.subs, markets); err != nil {
		return err
	}

	if err := a.withAuth(func() {
		a.writechn <- newWebSocketMessage(actionUnsubscribe, channelNameAccount, markets)
	}); err != nil {
		return err
	}

	return a.deleteSubscriptions(a.subs, a.closeInChannels(a.subs, markets), a.countSubscriptions(a.subs))
}

func (a *accountEventHandler) UnsubscribeAll() error {
	if err := a.Unsubscribe(a.subs.Keys()); err != nil {
		return err
	}

	return nil
}

func (a *accountEventHandler) handleOrderMessage(bytes []byte) {
	var orderEvent *OrderEvent
	if err := json.Unmarshal(bytes, &orderEvent); err != nil {
		log.Err(err).Str("message", string(bytes)).Msg("Couldn't unmarshal message into OrderEvent")
	} else {
		market := orderEvent.Market
		sub, exist := a.subs.Get(market)
		if exist {
			sub.orderinchn <- *orderEvent
		} else {
			log.Error().Str("market", market).Msg("There is no active subscription to handle this OrderEvent")
		}
	}
}

func (a *accountEventHandler) handleFillMessage(bytes []byte) {
	var fillEvent *FillEvent
	if err := json.Unmarshal(bytes, &fillEvent); err != nil {
		log.Err(err).Str("message", string(bytes)).Msg("Couldn't unmarshal message into FillEvent")
	} else {
		market := fillEvent.Market
		sub, exist := a.subs.Get(market)
		if exist {
			sub.fillinchn <- *fillEvent
		} else {
			log.Error().Str("market", market).Msg("There is no active subscription to handle this FillEvent")
		}
	}
}

func (a *accountEventHandler) handleAuthMessage(bytes []byte) {
	var authEvent *AuthEvent
	if err := json.Unmarshal(bytes, &authEvent); err != nil {
		log.Err(err).Str("message", string(bytes)).Msg("Couldn't unmarshal message into AuthEvent")
		a.authchn <- false
	} else {
		a.authchn <- authEvent.Authenticated
	}
}

func newWebSocketAuthMessage(apiKey string, apiSecret string) WebSocketMessage {
	timestamp := time.Now().UnixMilli()
	return WebSocketMessage{
		Action:    actionAuthenticate.Value,
		Key:       apiKey,
		Signature: crypto.CreateSignature("GET", "/websocket", nil, timestamp, apiSecret),
		Timestamp: timestamp,
	}
}

func (a *accountEventHandler) authenticate() {
	a.writechn <- newWebSocketAuthMessage(a.apiKey, a.apiSecret)
	a.authenticated = <-a.authchn
}

func (a *accountEventHandler) reconnect() {
	a.authenticated = false

	if err := a.withAuth(func() {
		a.writechn <- newWebSocketMessage(actionSubscribe, channelNameAccount, a.subs.Keys())
	}); err != nil {
		log.Err(err).Msg("Failed to reconnect the account websocket")
	}
}

func (a *accountEventHandler) withAuth(action func()) error {
	if !a.authenticated {
		a.authenticate()
	}

	if a.authenticated {
		action()
		return nil
	}

	return errAuthenticationFailed
}

func (a *accountEventHandler) closeInChannels(subs *safemap.SafeMap[string, *accountSubscription], markets []string) map[uuid.UUID][]string {
	idsWithMarkets := make(map[uuid.UUID][]string)
	for _, key := range markets {
		if sub, found := subs.Get(key); found {
			idsWithMarkets[sub.id] = append(idsWithMarkets[sub.id], key)
			close(sub.orderinchn)
			close(sub.fillinchn)
		}
	}
	return idsWithMarkets
}

func (a *accountEventHandler) deleteSubscriptions(
	subs *safemap.SafeMap[string, *accountSubscription],
	idsWithMarkets map[uuid.UUID][]string,
	idsWithCount map[uuid.UUID]int,
) error {
	for id, key := range idsWithMarkets {
		if idsWithCount[id] == len(key) {
			if item, found := subs.Get(key[0]); found {
				close(item.orderoutchn)
				close(item.filloutchn)
			}
		}
		for _, key := range key {
			subs.Remove(key)
		}
	}

	return nil
}

func (a *accountEventHandler) countSubscriptions(subs *safemap.SafeMap[string, *accountSubscription]) map[uuid.UUID]int {
	idsWithCount := make(map[uuid.UUID]int)
	for item := range subs.IterBuffered() {
		idsWithCount[item.Val.id]++
	}
	return idsWithCount
}
