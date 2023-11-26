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

	// The current status of the order.
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

	// Only for stop orders.
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
	// Describes the returned event over the socket.
	Event string `json:"event"`

	// The market which was requested in the subscription.
	Market string `json:"market"`

	// The order itself.
	Order Order `json:"order"`
}

func (o *OrderEvent) UnmarshalJSON(data []byte) error {
	var orderEvent map[string]any
	err := json.Unmarshal(data, &orderEvent)
	if err != nil {
		return err
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
		triggerPrice     = getOrEmpty("triggerPrice", orderEvent)
		triggerAmount    = getOrEmpty("triggerAmount", orderEvent)
		triggerType      = getOrEmpty("triggerType", orderEvent)
		triggerReference = getOrEmpty("triggerReference", orderEvent)
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

type Fill struct {
	// The id of the order on which has been filled
	OrderId string `json:"orderId"`
	// The id of the returned fill
	FillId string `json:"fillId"`
	// The current timestamp in milliseconds since 1 Jan 1970
	Timestamp int64 `json:"timestamp"`
	// The amount in base currency for which the trade has been made
	Amount float64 `json:"amount"`
	// The side for the taker
	// Enum: "buy" | "sell"
	Side string `json:"side"`
	// The price in quote currency for which the trade has been made
	Price float64 `json:"price"`
	// True for takers, false for makers
	Taker bool `json:"taker"`
	// The amount of fee that has been paid. Value is negative for rebates. Only available if settled is true
	Fee float64 `json:"fee"`
	// Currency in which the fee has been paid. Only available if settled is true
	FeeCurrency string `json:"feeCurrency"`
}

type FillEvent struct {
	// Describes the returned event over the socket
	Event string `json:"event"`
	// The market which was requested in the subscription
	Market string `json:"market"`
	// The fill itself
	Fill Fill `json:"fill"`
}

func (f *FillEvent) UnmarshalJSON(data []byte) error {
	var fillEvent map[string]any
	err := json.Unmarshal(data, &fillEvent)
	if err != nil {
		return err
	}

	var (
		market    = fillEvent["market"].(string)
		event     = fillEvent["event"].(string)
		orderId   = fillEvent["orderId"].(string)
		fillId    = fillEvent["fillId"].(string)
		timestamp = fillEvent["timestamp"].(float64)
		amount    = fillEvent["amount"].(string)
		side      = fillEvent["side"].(string)
		price     = fillEvent["price"].(string)
		taker     = fillEvent["taker"].(bool)

		// only available if settled is true
		fee         = getOrEmpty("fee", fillEvent)
		feeCurrency = getOrEmpty("feeCurrency", fillEvent)
	)

	f.Market = market
	f.Event = event
	f.Fill = Fill{
		OrderId:     orderId,
		FillId:      fillId,
		Timestamp:   int64(timestamp),
		Amount:      util.IfOrElse(len(amount) > 0, func() float64 { return util.MustFloat64(amount) }, ZERO),
		Side:        side,
		Price:       util.IfOrElse(len(price) > 0, func() float64 { return util.MustFloat64(price) }, ZERO),
		Taker:       taker,
		Fee:         util.IfOrElse(len(fee) > 0, func() float64 { return util.MustFloat64(fee) }, ZERO),
		FeeCurrency: feeCurrency,
	}

	return nil
}

type AccountSub interface {
	// Order channel to receive order events
	// You can set the buffSize for this channel, 0 for no buffer
	Order(buffSize uint64) <-chan OrderEvent

	// Order channel to receive fill events
	// You can set the buffSize for this channel, 0 for no buffer
	Fill(buffSize uint64) <-chan FillEvent
}

type accountSub struct {
	orderchn chan<- OrderEvent
	fillchn  chan<- FillEvent
}

func (a *accountSub) Order(buffSize uint64) <-chan OrderEvent {
	orderchn := make(chan OrderEvent, buffSize)
	a.orderchn = orderchn
	return orderchn
}

func (a *accountSub) Fill(buffSize uint64) <-chan FillEvent {
	fillchn := make(chan FillEvent, buffSize)
	a.fillchn = fillchn
	return fillchn
}

type AccountEventHandler interface {
	// Subscribe to market
	Subscribe(market string) (AccountSub, error)

	// Unsubscribe from market
	Unsubscribe(market string) error

	// Unsubscribe from every market
	UnsubscribeAll() error
}

type accountEventHandler struct {
	apiKey        string
	apiSecret     string
	windowTimeMs  uint64
	authenticated bool
	authchn       chan bool
	writechn      chan<- WebSocketMessage
	subs          *safemap.SafeMap[string, *accountSub]
}

func newAccountEventHandler(apiKey string, apiSecret string, windowTimeMs uint64, writechn chan<- WebSocketMessage) *accountEventHandler {
	return &accountEventHandler{
		apiKey:       apiKey,
		apiSecret:    apiSecret,
		windowTimeMs: windowTimeMs,
		writechn:     writechn,
		authchn:      make(chan bool),
		subs:         safemap.New[string, *accountSub](),
	}
}

func (t *accountEventHandler) Subscribe(market string) (AccountSub, error) {
	if t.subs.Has(market) {
		return nil, fmt.Errorf("subscription already active for market: %s", market)
	}

	if err := t.withAuth(func() {
		t.writechn <- newWebSocketMessage(ActionSubscribe, ChannelNameAccount, market)
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
		log.Logger().Error("Couldn't unmarshal message into OrderEvent", "message", string(bytes))
	} else if t.hasOrderChn(orderEvent.Market) {
		sub, _ := t.subs.Get(orderEvent.Market)
		sub.orderchn <- *orderEvent
	}
}

func (t *accountEventHandler) handleFillMessage(bytes []byte) {
	var fillEvent *FillEvent
	if err := json.Unmarshal(bytes, &fillEvent); err != nil {
		log.Logger().Error("Couldn't unmarshal message into FillEvent", "message", string(bytes))
	} else if t.hasFillChn(fillEvent.Market) {
		sub, _ := t.subs.Get(fillEvent.Market)
		sub.fillchn <- *fillEvent
	}
}

func (t *accountEventHandler) handleAuthMessage(bytes []byte) {
	var authEvent *AuthEvent
	if err := json.Unmarshal(bytes, &authEvent); err != nil {
		log.Logger().Error("Couldn't unmarshal message into AuthEvent", "message", string(bytes))
		t.authchn <- false
	} else {
		t.authchn <- authEvent.Authenticated
	}
}

func newWebSocketAuthMessage(apiKey string, apiSecret string, windowTimeMs uint64) WebSocketMessage {
	timestamp := time.Now().UnixMilli()
	return WebSocketMessage{
		Action:    ActionAuthenticate.Value,
		Key:       apiKey,
		Signature: createSignature(timestamp, apiSecret),
		Timestamp: timestamp,
		Window:    windowTimeMs,
	}
}

func createSignature(timestamp int64, apiSecret string) string {
	hash := hmac.New(sha256.New, []byte(apiSecret))
	hash.Write([]byte(fmt.Sprintf("%dGET/v2/websocket", timestamp)))
	return hex.EncodeToString(hash.Sum(nil))
}

func (t *accountEventHandler) authenticate() {
	t.writechn <- newWebSocketAuthMessage(t.apiKey, t.apiSecret, t.windowTimeMs)
	t.authenticated = <-t.authchn
}

func (t *accountEventHandler) reconnect() {
	t.authenticated = false

	for sub := range t.subs.IterBuffered() {
		market := sub.Key
		if err := t.withAuth(func() {
			t.writechn <- newWebSocketMessage(ActionSubscribe, ChannelNameAccount, market)
		}); err != nil {
			log.Logger().Error("Failed to reconnect the account websocket", "market", market)
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

	return fmt.Errorf("could not subscribe, authentication failed")
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

func getOrEmpty(key string, data map[string]any) string {
	value, exist := data[key]
	return util.IfOrElse(exist, func() string { return value.(string) }, "")
}
