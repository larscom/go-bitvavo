package ws

import (
	"time"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/larscom/go-bitvavo/v2/crypto"
	"github.com/larscom/go-bitvavo/v2/util"
	"github.com/rs/zerolog/log"

	"github.com/larscom/go-bitvavo/v2/types"
	csmap "github.com/mhmtszr/concurrent-swiss-map"
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
	subs          *csmap.CsMap[string, *accountSubscription]
}

func newAccountEventHandler(apiKey string, apiSecret string, writechn chan<- WebSocketMessage) *accountEventHandler {
	return &accountEventHandler{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		writechn:  writechn,
		authchn:   make(chan bool),
		subs:      csmap.Create[string, *accountSubscription](),
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
		orderoutchn = make(chan OrderEvent, int(size)*len(markets))
		filloutchn  = make(chan FillEvent, int(size)*len(markets))
		id          = uuid.New()
	)

	for _, market := range markets {
		orderinchn := make(chan OrderEvent, size)
		fillinchn := make(chan FillEvent, size)

		a.subs.Store(market, newAccountSubscription(id, market, orderinchn, orderoutchn, fillinchn, filloutchn))

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

	return a.deleteSubscriptions(a.subs, markets)
}

func (a *accountEventHandler) UnsubscribeAll() error {
	if err := a.Unsubscribe(getSubscriptionKeys(a.subs)); err != nil {
		return err
	}

	return nil
}

func (a *accountEventHandler) handleMessage(e WsEvent, bytes []byte) {
	switch e {
	case wsEventAuth:
		a.handleAuthMessage(bytes)
	case wsEventOrder:
		a.handleOrderMessage(bytes)
	case wsEventFill:
		a.handleFillMessage(bytes)
	default:
		log.Debug().Str("event", e.Value).Msg("no handler for this account event (should not happen)")
	}
}

func (a *accountEventHandler) handleOrderMessage(bytes []byte) {
	var orderEvent *OrderEvent
	if err := json.Unmarshal(bytes, &orderEvent); err != nil {
		log.Err(err).Str("message", string(bytes)).Msg("Couldn't unmarshal message into OrderEvent")
	} else {
		market := orderEvent.Market
		sub, exist := a.subs.Load(market)
		if exist {
			sub.orderinchn <- *orderEvent
		} else {
			log.Debug().Str("market", market).Msg("There is no active subscription to handle this OrderEvent")
		}
	}
}

func (a *accountEventHandler) handleFillMessage(bytes []byte) {
	var fillEvent *FillEvent
	if err := json.Unmarshal(bytes, &fillEvent); err != nil {
		log.Err(err).Str("message", string(bytes)).Msg("Couldn't unmarshal message into FillEvent")
	} else {
		market := fillEvent.Market
		sub, exist := a.subs.Load(market)
		if exist {
			sub.fillinchn <- *fillEvent
		} else {
			log.Debug().Str("market", market).Msg("There is no active subscription to handle this FillEvent")
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
		a.writechn <- newWebSocketMessage(actionSubscribe, channelNameAccount, getSubscriptionKeys(a.subs))
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

func (a *accountEventHandler) deleteSubscriptions(
	subs *csmap.CsMap[string, *accountSubscription],
	markets []string,
) error {
	counts := make(map[uuid.UUID]int)
	subs.Range(func(key string, value *accountSubscription) (stop bool) {
		counts[value.id]++
		return false
	})

	idsWithKeys := make(map[uuid.UUID][]string)
	for _, key := range markets {
		if sub, found := subs.Load(key); found {
			idsWithKeys[sub.id] = append(idsWithKeys[sub.id], key)
			close(sub.orderinchn)
			close(sub.fillinchn)
		}
	}

	for id, keys := range idsWithKeys {
		if counts[id] == len(keys) {
			if item, found := subs.Load(keys[0]); found {
				close(item.orderoutchn)
				close(item.filloutchn)
			}
		}
		for _, key := range keys {
			subs.Delete(key)
		}
	}

	return nil
}
