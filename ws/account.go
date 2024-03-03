package ws

import (
	"log/slog"
	"time"

	"github.com/goccy/go-json"
	"github.com/larscom/go-bitvavo/v2/crypto"

	"github.com/larscom/go-bitvavo/v2/types"
	"github.com/larscom/go-bitvavo/v2/util"
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
	err := json.Unmarshal(bytes, &orderEvent)
	if err != nil {
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
	err := json.Unmarshal(bytes, &fillEvent)
	if err != nil {
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

type AccountSubscription interface {
	// Order channel to receive order events.
	// You can set the buffSize for this channel.
	//
	// If you have many subscriptions at once you may need to increase the buffSize
	//
	// Default buffSize: 50
	Order(buffSize ...uint64) <-chan OrderEvent

	// Order channel to receive fill events.
	// You can set the buffSize for this channel.
	//
	// If you have many subscriptions at once you may need to increase the buffSize
	//
	// Default buffSize: 50
	Fill(buffSize ...uint64) <-chan FillEvent
}

type accountSub struct {
	orderchn chan<- OrderEvent
	fillchn  chan<- FillEvent
}

func (a *accountSub) Order(buffSize ...uint64) <-chan OrderEvent {
	size := util.IfOrElse(len(buffSize) > 0, func() uint64 { return buffSize[0] }, defaultBuffSize)

	orderchn := make(chan OrderEvent, size)
	a.orderchn = orderchn

	return orderchn
}

func (a *accountSub) Fill(buffSize ...uint64) <-chan FillEvent {
	size := util.IfOrElse(len(buffSize) > 0, func() uint64 { return buffSize[0] }, defaultBuffSize)

	fillchn := make(chan FillEvent, size)
	a.fillchn = fillchn

	return fillchn
}

type AccountEventHandler interface {
	// Subscribe to market
	Subscribe(market string) (AccountSubscription, error)

	// Unsubscribe from market
	Unsubscribe(market string) error

	// Unsubscribe from every market
	UnsubscribeAll() error
}

type accountEventHandler struct {
	apiKey        string
	apiSecret     string
	authenticated bool
	authchn       chan bool
	writechn      chan<- WebSocketMessage
	subs          *safemap.SafeMap[string, *accountSub]
}

func newAccountEventHandler(apiKey string, apiSecret string, writechn chan<- WebSocketMessage) *accountEventHandler {
	return &accountEventHandler{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		writechn:  writechn,
		authchn:   make(chan bool),
		subs:      safemap.New[string, *accountSub](),
	}
}

func (t *accountEventHandler) Subscribe(market string) (AccountSubscription, error) {
	if t.subs.Has(market) {
		return nil, errSubscriptionAlreadyActive
	}

	if err := t.withAuth(func() {
		t.writechn <- newWebSocketMessage(actionSubscribe, channelNameAccount, market)
	}); err != nil {
		return nil, err
	}

	subscription := new(accountSub)

	t.subs.Set(market, subscription)

	return subscription, nil

}

func (t *accountEventHandler) Unsubscribe(market string) error {
	sub, exist := t.subs.Get(market)

	if exist {
		if err := t.withAuth(func() {
			t.writechn <- newWebSocketMessage(actionUnsubscribe, channelNameBook, market)
		}); err != nil {
			return err
		}
		if sub.fillchn != nil {
			close(sub.fillchn)
		}
		if sub.orderchn != nil {
			close(sub.orderchn)
		}
		t.subs.Remove(market)
		return nil
	}

	return errNoSubscriptionActive
}

func (t *accountEventHandler) UnsubscribeAll() error {
	for sub := range t.subs.IterBuffered() {
		market := sub.Key
		if err := t.Unsubscribe(market); err != nil {
			return err
		}
	}
	return nil
}

func (t *accountEventHandler) handleOrderMessage(bytes []byte) {
	var orderEvent *OrderEvent
	if err := json.Unmarshal(bytes, &orderEvent); err != nil {
		slog.Error("Couldn't unmarshal message into OrderEvent", "message", string(bytes))
	} else if t.hasOrderChn(orderEvent.Market) {
		sub, _ := t.subs.Get(orderEvent.Market)
		sub.orderchn <- *orderEvent
	}
}

func (t *accountEventHandler) handleFillMessage(bytes []byte) {
	var fillEvent *FillEvent
	if err := json.Unmarshal(bytes, &fillEvent); err != nil {
		slog.Error("Couldn't unmarshal message into FillEvent", "message", string(bytes))
	} else if t.hasFillChn(fillEvent.Market) {
		sub, _ := t.subs.Get(fillEvent.Market)
		sub.fillchn <- *fillEvent
	}
}

func (t *accountEventHandler) handleAuthMessage(bytes []byte) {
	var authEvent *AuthEvent
	if err := json.Unmarshal(bytes, &authEvent); err != nil {
		slog.Error("Couldn't unmarshal message into AuthEvent", "message", string(bytes))
		t.authchn <- false
	} else {
		t.authchn <- authEvent.Authenticated
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

func (t *accountEventHandler) authenticate() {
	t.writechn <- newWebSocketAuthMessage(t.apiKey, t.apiSecret)
	t.authenticated = <-t.authchn
}

func (t *accountEventHandler) reconnect() {
	t.authenticated = false

	for sub := range t.subs.IterBuffered() {
		market := sub.Key
		if err := t.withAuth(func() {
			t.writechn <- newWebSocketMessage(actionSubscribe, channelNameAccount, market)
		}); err != nil {
			slog.Error("Failed to reconnect the account websocket", "market", market)
		}
	}
}

func (t *accountEventHandler) withAuth(action func()) error {
	if !t.authenticated {
		t.authenticate()
	}

	if t.authenticated {
		action()
		return nil
	}

	return errAuthenticationFailed
}

func (t *accountEventHandler) hasOrderChn(market string) bool {
	sub, exist := t.subs.Get(market)

	if exist {
		return sub.orderchn != nil
	}

	return false
}

func (t *accountEventHandler) hasFillChn(market string) bool {
	sub, exist := t.subs.Get(market)

	if exist {
		return sub.fillchn != nil
	}

	return false
}
