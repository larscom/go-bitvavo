package bitvavo

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/goccy/go-json"
	"github.com/larscom/go-bitvavo/v2/log"
	"github.com/larscom/go-bitvavo/v2/util"
	"github.com/smallnest/safemap"
)

type Order struct {
	Guid string `json:"guid"`
	// The order id of the returned order.
	OrderId string `json:"orderId"`
	// Is a timestamp in milliseconds since 1 Jan 1970.
	Created int64 `json:"created"`
	// Is a timestamp in milliseconds since 1 Jan 1970.
	Updated int64 `json:"updated"`
	// The current status of the order
	// Enum: "new" | "awaitingTrigger" | "canceled" | "canceledAuction" | "canceledSelfTradePrevention" | "canceledIOC" | "canceledFOK" | "canceledMarketProtection" | "canceledPostOnly" | "filled" | "partiallyFilled" | "expired" | "rejected"
	Status string `json:"status"`
	// Side
	// Enum: "buy" | "sell"
	Side string `json:"side"`
	// OrderType
	// Enum: "limit" | "market"
	OrderType string `json:"orderType"`
	// Original amount.
	Amount float64 `json:"amount"`
	// Amount remaining (lower than 'amount' after fills).
	AmountRemaining float64 `json:"amountRemaining"`
	// The price of the order.
	Price float64 `json:"price"`
	// Amount of 'onHoldCurrency' that is reserved for this order. This is released when orders are canceled.
	OnHold float64 `json:"onHold"`
	// The currency placed on hold is the quote currency for sell orders and base currency for buy orders.
	OnHoldCurrency string `json:"onHoldCurrency"`
	// Only for stop orders: The current price used in the trigger. This is based on the triggerAmount and triggerType.
	TriggerPrice float64 `json:"triggerPrice"`
	// Only for stop orders: The value used for the triggerType to determine the triggerPrice.
	TriggerAmount float64 `json:"triggerAmount"`
	// Only for stop orders
	// Enum: "price"
	TriggerType string `json:"triggerType"`
	// Only for stop orders: The reference price used for stop orders.
	// Enum: "lastTrade" | "bestBid" | "bestAsk" | "midPrice"
	TriggerReference string `json:"triggerReference"`
	// Only for limit orders: Determines how long orders remain active.
	// Possible values: Good-Til-Canceled (GTC), Immediate-Or-Cancel (IOC), Fill-Or-Kill (FOK).
	// GTC orders will remain on the order book until they are filled or canceled.
	// IOC orders will fill against existing orders, but will cancel any remaining amount after that.
	// FOK orders will fill against existing orders in its entirety, or will be canceled (if the entire order cannot be filled).
	// Enum: "GTC" | "IOC" | "FOK"
	TimeInForce string `json:"timeInForce"`
	// Default: false
	PostOnly bool `json:"postOnly"`
	// Self trading is not allowed on Bitvavo. Multiple options are available to prevent this from happening.
	// The default ‘decrementAndCancel’ decrements both orders by the amount that would have been filled, which in turn cancels the smallest of the two orders.
	// ‘cancelOldest’ will cancel the entire older order and places the new order.
	// ‘cancelNewest’ will cancel the order that is submitted.
	// ‘cancelBoth’ will cancel both the current and the old order.
	// Default: "decrementAndCancel"
	// Enum: "decrementAndCancel" | "cancelOldest" | "cancelNewest" | "cancelBoth"
	SelfTradePrevention string `json:"selfTradePrevention"`
	// Whether this order is visible on the order book.
	Visible bool `json:"visible"`
}

type OrderEvent struct {
	// Describes the returned event over the socket
	Event string `json:"event"`
	// The market which was requested in the subscription
	Market string `json:"market"`
	// The order itself
	Order Order `json:"order"`
}

func (o *OrderEvent) UnmarshalJSON(data []byte) error {
	var orderEvent map[string]any
	err := json.Unmarshal(data, &orderEvent)
	if err != nil {
		return err
	}

	getOrEmpty := func(key string) string {
		value, exist := orderEvent[key]
		return util.IfOrElse(exist, func() string { return value.(string) }, "")
	}

	var (
		market              = orderEvent["market"].(string)
		event               = orderEvent["event"].(string)
		guid                = orderEvent["guid"].(string)
		orderId             = orderEvent["orderId"].(string)
		created             = orderEvent["created"].(float64)
		updated             = orderEvent["updated"].(float64)
		status              = orderEvent["status"].(string)
		side                = orderEvent["side"].(string)
		orderType           = orderEvent["orderType"].(string)
		amount              = orderEvent["amount"].(string)
		amountRemaining     = orderEvent["amountRemaining"].(string)
		price               = orderEvent["price"].(string)
		onHold              = orderEvent["onHold"].(string)
		onHoldCurrency      = orderEvent["onHoldCurrency"].(string)
		timeInForce         = orderEvent["timeInForce"].(string)
		postOnly            = orderEvent["postOnly"].(bool)
		selfTradePrevention = orderEvent["selfTradePrevention"].(string)
		visible             = orderEvent["visible"].(bool)

		// only for stop orders
		triggerPrice     = getOrEmpty("triggerPrice")
		triggerAmount    = getOrEmpty("triggerAmount")
		triggerType      = getOrEmpty("triggerType")
		triggerReference = getOrEmpty("triggerReference")
	)

	o.Market = market
	o.Event = event
	o.Order = Order{
		Guid:                guid,
		OrderId:             orderId,
		Created:             int64(created),
		Updated:             int64(updated),
		Status:              status,
		Side:                side,
		OrderType:           orderType,
		Amount:              util.IfOrElse(len(amount) > 0, func() float64 { return util.MustFloat64(amount) }, ZERO),
		AmountRemaining:     util.IfOrElse(len(amountRemaining) > 0, func() float64 { return util.MustFloat64(amountRemaining) }, ZERO),
		Price:               util.IfOrElse(len(price) > 0, func() float64 { return util.MustFloat64(price) }, ZERO),
		OnHold:              util.IfOrElse(len(onHold) > 0, func() float64 { return util.MustFloat64(onHold) }, ZERO),
		OnHoldCurrency:      onHoldCurrency,
		TriggerPrice:        util.IfOrElse(len(triggerPrice) > 0, func() float64 { return util.MustFloat64(triggerPrice) }, ZERO),
		TriggerAmount:       util.IfOrElse(len(triggerAmount) > 0, func() float64 { return util.MustFloat64(triggerAmount) }, ZERO),
		TriggerType:         triggerType,
		TriggerReference:    triggerReference,
		TimeInForce:         timeInForce,
		PostOnly:            postOnly,
		SelfTradePrevention: selfTradePrevention,
		Visible:             visible,
	}

	return nil
}

type Fill struct{}

type FillEvent struct {
	// Describes the returned event over the socket
	Event string `json:"event"`
	// The market which was requested in the subscription
	Market string `json:"market"`
	// The fill itself
	Fill Fill `json:"fill"`
}

type AccountSubscription interface {
	// Order channel to receive order events
	// You can set the buffSize for this channel, 0 for no buffer
	Order(buffSize uint64) <-chan OrderEvent
	// Order channel to receive fill events
	// You can set the buffSize for this channel, 0 for no buffer
	Fill(buffSize uint64) <-chan FillEvent
}

type accountSubscription struct {
	orderchn chan<- OrderEvent
	fillchn  chan<- FillEvent
}

func (a *accountSubscription) Order(buffSize uint64) <-chan OrderEvent {
	orderchn := make(chan OrderEvent, buffSize)
	a.orderchn = orderchn
	return orderchn
}

func (a *accountSubscription) Fill(buffSize uint64) <-chan FillEvent {
	fillchn := make(chan FillEvent, buffSize)
	a.fillchn = fillchn
	return fillchn
}

type AccountWsHandler interface {
	// Subscribe to market
	Subscribe(market string) (AccountSubscription, error)

	// Unsubscribe from market
	Unsubscribe(market string) error

	// Unsubscribe from every market
	UnsubscribeAll() error
}

type accountWsHandler struct {
	apiKey        string
	apiSecret     string
	authenticated bool
	authchn       chan bool
	writechn      chan<- WebSocketMessage
	subs          *safemap.SafeMap[string, *accountSubscription]
}

func newAccountWsHandler(apiKey string, apiSecret string, writechn chan<- WebSocketMessage) *accountWsHandler {
	return &accountWsHandler{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		writechn:  writechn,
		authchn:   make(chan bool),
		subs:      safemap.New[string, *accountSubscription](),
	}
}

func (t *accountWsHandler) Subscribe(market string) (AccountSubscription, error) {
	if t.subs.Has(market) {
		return nil, fmt.Errorf("subscription already active for market: %s", market)
	}

	if err := t.withAuth(func() {
		t.writechn <- newWebSocketMessage(ActionSubscribe, ChannelNameAccount, market)
	}); err != nil {
		return nil, err
	}

	subscription := new(accountSubscription)

	t.subs.Set(market, subscription)

	return subscription, nil

}

func (t *accountWsHandler) Unsubscribe(market string) error {
	sub, exist := t.subs.Get(market)

	if exist {
		if err := t.withAuth(func() {
			t.writechn <- newWebSocketMessage(ActionUnsubscribe, ChannelNameBook, market)
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

	return fmt.Errorf("no subscription active for market: %s", market)
}

func (t *accountWsHandler) UnsubscribeAll() error {
	for sub := range t.subs.IterBuffered() {
		market := sub.Key
		if err := t.Unsubscribe(market); err != nil {
			return err
		}
	}
	return nil
}

func (t *accountWsHandler) handleOrderMessage(bytes []byte) {
	var orderEvent *OrderEvent
	if err := json.Unmarshal(bytes, &orderEvent); err != nil {
		log.Logger().Error("Couldn't unmarshal message into OrderEvent", "message", string(bytes))
	} else if t.hasOrderChn(orderEvent.Market) {
		sub, _ := t.subs.Get(orderEvent.Market)
		sub.orderchn <- *orderEvent
	}
}

func (t *accountWsHandler) handleFillMessage(bytes []byte) {
	var fillEvent *FillEvent
	if err := json.Unmarshal(bytes, &fillEvent); err != nil {
		log.Logger().Error("Couldn't unmarshal message into FillEvent", "message", string(bytes))
	} else if t.hasFillChn(fillEvent.Market) {
		sub, _ := t.subs.Get(fillEvent.Market)
		sub.fillchn <- *fillEvent
	}
}

func (t *accountWsHandler) handleAuthMessage(bytes []byte) {
	var authEvent *AuthEvent
	if err := json.Unmarshal(bytes, &authEvent); err != nil {
		log.Logger().Error("Couldn't unmarshal message into AuthEvent", "message", string(bytes))
		t.authchn <- false
	} else {
		t.authchn <- authEvent.Authenticated
	}
}

func newWebSocketAuthMessage(apiKey string, apiSecret string) WebSocketMessage {
	timestamp := time.Now().UnixMilli()
	return WebSocketMessage{
		Action:    ActionAuthenticate.Value,
		Key:       apiKey,
		Signature: createSignature(timestamp, apiSecret),
		Timestamp: timestamp,
		Window:    10000,
	}
}

func createSignature(timestamp int64, apiSecret string) string {
	hash := hmac.New(sha256.New, []byte(apiSecret))
	hash.Write([]byte(fmt.Sprintf("%dGET/v2/websocket", timestamp)))
	sha := hex.EncodeToString(hash.Sum(nil))
	return sha
}

func (t *accountWsHandler) authenticate() {
	t.writechn <- newWebSocketAuthMessage(t.apiKey, t.apiSecret)
	t.authenticated = <-t.authchn
}

func (t *accountWsHandler) reconnect() {
	t.authenticated = false

	for sub := range t.subs.IterBuffered() {
		market := sub.Key
		t.Subscribe(market)
	}
}

func (t *accountWsHandler) withAuth(action func()) error {
	if !t.authenticated {
		t.authenticate()
	}

	if t.authenticated {
		action()
		return nil
	}

	return fmt.Errorf("could not subscribe, authentication failed")
}

func (t *accountWsHandler) hasOrderChn(market string) bool {
	sub, exist := t.subs.Get(market)

	if exist {
		return sub.orderchn != nil
	}

	return false
}

func (t *accountWsHandler) hasFillChn(market string) bool {
	sub, exist := t.subs.Get(market)

	if exist {
		return sub.fillchn != nil
	}

	return false
}
