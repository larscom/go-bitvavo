package types

import (
	"fmt"
	"net/url"
	"time"

	"github.com/goccy/go-json"
	"github.com/larscom/go-bitvavo/v2/util"
)

type OrderParams struct {
	// Return the limit most recent orders only.
	// Default: 500
	Limit uint64 `json:"limit"`

	// Return orders after start time.
	Start time.Time `json:"start"`

	// Return orders before end time.
	End time.Time `json:"end"`

	// Filter used to limit the returned results.
	// All orders after this order ID are returned (i.e. showing those later in time).
	OrderIdFrom string `json:"orderIdFrom"`

	// Filter used to limit the returned results.
	// All orders up to this order ID are returned (i.e. showing those earlier in time).
	OrderIdTo string `json:"orderIdTo"`
}

func (o *OrderParams) Params() url.Values {
	params := make(url.Values)
	if o.Limit > 0 {
		params.Add("limit", fmt.Sprint(o.Limit))
	}
	if !o.Start.IsZero() {
		params.Add("start", fmt.Sprint(o.Start.UnixMilli()))
	}
	if !o.End.IsZero() {
		params.Add("end", fmt.Sprint(o.End.UnixMilli()))
	}
	if o.OrderIdFrom != "" {
		params.Add("orderIdFrom", o.OrderIdFrom)
	}
	if o.OrderIdTo != "" {
		params.Add("orderIdTo", o.OrderIdTo)
	}
	return params
}

type Order struct {
	// The order id of the returned order.
	OrderId string `json:"orderId"`

	// The personalized UUID for this orderId in this market.
	ClientOrderId string `json:"clientOrderId"`

	// The market in which the order was placed.
	Market string `json:"market"`

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

func (o *Order) UnmarshalJSON(bytes []byte) error {
	var j map[string]any
	err := json.Unmarshal(bytes, &j)
	if err != nil {
		return err
	}

	var (
		orderId             = j["orderId"].(string)
		market              = j["market"].(string)
		created             = j["created"].(float64)
		updated             = j["updated"].(float64)
		status              = j["status"].(string)
		side                = j["side"].(string)
		orderType           = j["orderType"].(string)
		amount              = j["amount"].(string)
		amountRemaining     = j["amountRemaining"].(string)
		price               = j["price"].(string)
		onHold              = j["onHold"].(string)
		onHoldCurrency      = j["onHoldCurrency"].(string)
		timeInForce         = j["timeInForce"].(string)
		postOnly            = j["postOnly"].(bool)
		selfTradePrevention = j["selfTradePrevention"].(string)
		visible             = j["visible"].(bool)

		clientOrderId = GetOrEmpty[string]("clientOrderId", j)

		// only for stop orders
		triggerPrice     = GetOrEmpty[string]("triggerPrice", j)
		triggerAmount    = GetOrEmpty[string]("triggerAmount", j)
		triggerType      = GetOrEmpty[string]("triggerType", j)
		triggerReference = GetOrEmpty[string]("triggerReference", j)
	)

	o.OrderId = orderId
	o.ClientOrderId = clientOrderId
	o.Market = market
	o.Created = int64(created)
	o.Updated = int64(updated)
	o.Status = status
	o.Side = side
	o.OrderType = orderType
	o.Amount = util.IfOrElse(len(amount) > 0, func() float64 { return util.MustFloat64(amount) }, 0)
	o.AmountRemaining = util.IfOrElse(len(amountRemaining) > 0, func() float64 { return util.MustFloat64(amountRemaining) }, 0)
	o.Price = util.IfOrElse(len(price) > 0, func() float64 { return util.MustFloat64(price) }, 0)
	o.OnHold = util.IfOrElse(len(onHold) > 0, func() float64 { return util.MustFloat64(onHold) }, 0)
	o.OnHoldCurrency = onHoldCurrency
	o.TriggerPrice = util.IfOrElse(len(triggerPrice) > 0, func() float64 { return util.MustFloat64(triggerPrice) }, 0)
	o.TriggerAmount = util.IfOrElse(len(triggerAmount) > 0, func() float64 { return util.MustFloat64(triggerAmount) }, 0)
	o.TriggerType = triggerType
	o.TriggerReference = triggerReference
	o.TimeInForce = timeInForce
	o.PostOnly = postOnly
	o.SelfTradePrevention = selfTradePrevention
	o.Visible = visible

	return nil
}
