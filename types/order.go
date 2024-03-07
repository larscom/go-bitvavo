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

type OrderNew struct {
	// The market in which the order should be placed (e.g: ETH-EUR)
	Market string `json:"market"`

	// When placing a buy order the base currency will be bought for the quote currency. When placing a sell order the base currency will be sold for the quote currency.
	//
	// Enum: "buy" | "sell"
	Side string `json:"side"`

	// For limit orders, amount and price are required. For market orders either amount or amountQuote is required.
	//
	// Enum: "market" | "limit" | "stopLoss" | "stopLossLimit" | "takeProfit" | "takeProfitLimit"
	OrderType string `json:"orderType"`

	// Specifies the amount of the base asset that will be bought/sold.
	Amount float64 `json:"amount,omitempty"`

	// Only for limit orders: Specifies the amount in quote currency that is paid/received for each unit of base currency.
	Price float64 `json:"price,omitempty"`

	// Only for market orders: If amountQuote is specified, [amountQuote] of the quote currency will be bought/sold for the best price available.
	AmountQuote float64 `json:"amountQuote,omitempty"`

	// Only for stop orders: Specifies the amount that is used with the triggerType.
	// Combine this parameter with triggerType and triggerReference to create the desired trigger.
	TriggerAmount float64 `json:"triggerAmount,omitempty"`

	// Only for stop orders: Only allows price for now. A triggerAmount of 4000 and a triggerType of price will generate a triggerPrice of 4000.
	// Combine this parameter with triggerAmount and triggerReference to create the desired trigger.
	//
	// Enum: "price"
	TriggerType string `json:"triggerType,omitempty"`

	// Only for stop orders: Use this to determine which parameter will trigger the order.
	// Combine this parameter with triggerAmount and triggerType to create the desired trigger.
	//
	// Enum: "lastTrade" | "bestBid" | "bestAsk" | "midPrice"
	TriggerReference string `json:"triggerReference,omitempty"`

	// Only for limit orders: Determines how long orders remain active.
	// Possible values: Good-Til-Canceled (GTC), Immediate-Or-Cancel (IOC), Fill-Or-Kill (FOK).
	// GTC orders will remain on the order book until they are filled or canceled.
	// IOC orders will fill against existing orders, but will cancel any remaining amount after that.
	// FOK orders will fill against existing orders in its entirety, or will be canceled (if the entire order cannot be filled).
	//
	// Enum: "GTC" | "IOC" | "FOK"
	// Default: "GTC"
	TimeInForce string `json:"timeInForce,omitempty"`

	// Self trading is not allowed on Bitvavo. Multiple options are available to prevent this from happening.
	// The default ‘decrementAndCancel’ decrements both orders by the amount that would have been filled, which in turn cancels the smallest of the two orders.
	// ‘cancelOldest’ will cancel the entire older order and places the new order.
	// ‘cancelNewest’ will cancel the order that is submitted.
	// ‘cancelBoth’ will cancel both the current and the old order.
	// Default: "decrementAndCancel"
	//
	// Enum: "decrementAndCancel" | "cancelOldest" | "cancelNewest" | "cancelBoth"
	// Default: "decrementAndCancel"
	SelfTradePrevention string `json:"selfTradePrevention,omitempty"`

	// Only for limit orders: When postOnly is set to true, the order will not fill against existing orders.
	// This is useful if you want to ensure you pay the maker fee. If the order would fill against existing orders, the entire order will be canceled.
	//
	// Default: false
	PostOnly bool `json:"postOnly,omitempty"`

	// Only for market orders: In order to protect clients from filling market orders with undesirable prices,
	// the remainder of market orders will be canceled once the next fill price is 10% worse than the best fill price (best bid/ask at first match).
	// If you wish to disable this protection, set this value to ‘true’.
	//
	// Default: false
	DisableMarketProtection bool `json:"disableMarketProtection,omitempty"`

	// If this is set to 'true', all order information is returned.
	// Set this to 'false' when only an acknowledgement of success or failure is required, this is faster.
	//
	// Default: true
	ResponseRequired bool `json:"responseRequired,omitempty"`
}

type OrderUpdate struct {
	// The market for which an order should be updated
	Market string `json:"market"`

	// The id of the order which should be updated
	OrderId string `json:"orderId"`

	// Updates amount to this value (and also changes amountRemaining accordingly).
	Amount float64 `json:"amount,omitempty"`

	// Only for market orders: If amountQuote is specified, [amountQuote] of the quote currency will be bought/sold for the best price available.
	AmountQuote float64 `json:"amountQuote,omitempty"`

	// Updates amountRemaining to this value (and also changes amount accordingly).
	AmountRemaining float64 `json:"amountRemaining,omitempty"`

	// Specifies the amount in quote currency that is paid/received for each unit of base currency.
	Price float64 `json:"price,omitempty"`

	// Only for stop orders: Specifies the amount that is used with the triggerType.
	// Combine this parameter with triggerType and triggerReference to create the desired trigger.
	TriggerAmount float64 `json:"triggerAmount,omitempty"`

	// Only for limit orders: Determines how long orders remain active.
	// Possible values: Good-Til-Canceled (GTC), Immediate-Or-Cancel (IOC), Fill-Or-Kill (FOK).
	// GTC orders will remain on the order book until they are filled or canceled.
	// IOC orders will fill against existing orders, but will cancel any remaining amount after that.
	// FOK orders will fill against existing orders in its entirety, or will be canceled (if the entire order cannot be filled).
	//
	// Enum: "GTC" | "IOC" | "FOK"
	// Default: "GTC"
	TimeInForce string `json:"timeInForce,omitempty"`

	// Self trading is not allowed on Bitvavo. Multiple options are available to prevent this from happening.
	// The default ‘decrementAndCancel’ decrements both orders by the amount that would have been filled, which in turn cancels the smallest of the two orders.
	// ‘cancelOldest’ will cancel the entire older order and places the new order.
	// ‘cancelNewest’ will cancel the order that is submitted.
	// ‘cancelBoth’ will cancel both the current and the old order.
	// Default: "decrementAndCancel"
	//
	// Enum: "decrementAndCancel" | "cancelOldest" | "cancelNewest" | "cancelBoth"
	// Default: "decrementAndCancel"
	SelfTradePrevention string `json:"selfTradePrevention,omitempty"`

	// Only for limit orders: When postOnly is set to true, the order will not fill against existing orders.
	// This is useful if you want to ensure you pay the maker fee. If the order would fill against existing orders, the entire order will be canceled.
	//
	// Default: false
	PostOnly bool `json:"postOnly,omitempty"`

	// If this is set to 'true', all order information is returned.
	// Set this to 'false' when only an acknowledgement of success or failure is required, this is faster.
	//
	// Default: true
	ResponseRequired bool `json:"responseRequired,omitempty"`
}

type Order struct {
	// The order id of the returned order.
	OrderId string `json:"orderId"`

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
	// Enum: "market" | "limit" | "stopLoss" | "stopLossLimit" | "takeProfit" | "takeProfitLimit"
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
	//
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
	//
	// Enum: "GTC" | "IOC" | "FOK"
	TimeInForce string `json:"timeInForce"`

	// Default: false
	PostOnly bool `json:"postOnly"`

	// Self trading is not allowed on Bitvavo. Multiple options are available to prevent this from happening.
	// The default ‘decrementAndCancel’ decrements both orders by the amount that would have been filled, which in turn cancels the smallest of the two orders.
	// ‘cancelOldest’ will cancel the entire older order and places the new order.
	// ‘cancelNewest’ will cancel the order that is submitted.
	// ‘cancelBoth’ will cancel both the current and the old order.
	//
	// Default: "decrementAndCancel"
	// Enum: "decrementAndCancel" | "cancelOldest" | "cancelNewest" | "cancelBoth"
	SelfTradePrevention string `json:"selfTradePrevention"`

	// Whether this order is visible on the order book.
	Visible bool `json:"visible"`

	// The fills for this order
	Fills []Fill `json:"fills"`

	// How much of this order is filled
	FilledAmount float64 `json:"filledAmount"`

	// How much of this order is filled in quote currency
	FilledAmountQuote float64 `json:"filledAmountQuote"`

	// The currency in which the fee is payed (e.g: EUR)
	FeeCurrency string `json:"feeCurrency"`

	// How much fee is payed
	FeePaid float64 `json:feePaid""`
}

func (o *Order) UnmarshalJSON(bytes []byte) error {
	var j map[string]any

	if err := json.Unmarshal(bytes, &j); err != nil {
		return err
	}

	var (
		orderId             = getOrEmpty[string]("orderId", j)
		market              = getOrEmpty[string]("market", j)
		created             = getOrEmpty[float64]("created", j)
		updated             = getOrEmpty[float64]("updated", j)
		status              = getOrEmpty[string]("status", j)
		side                = getOrEmpty[string]("side", j)
		orderType           = getOrEmpty[string]("orderType", j)
		amount              = getOrEmpty[string]("amount", j)
		amountRemaining     = getOrEmpty[string]("amountRemaining", j)
		price               = getOrEmpty[string]("price", j)
		onHold              = getOrEmpty[string]("onHold", j)
		onHoldCurrency      = getOrEmpty[string]("onHoldCurrency", j)
		timeInForce         = getOrEmpty[string]("timeInForce", j)
		postOnly            = getOrEmpty[bool]("postOnly", j)
		selfTradePrevention = getOrEmpty[string]("selfTradePrevention", j)
		visible             = getOrEmpty[bool]("visible", j)

		// only for stop orders
		triggerPrice     = getOrEmpty[string]("triggerPrice", j)
		triggerAmount    = getOrEmpty[string]("triggerAmount", j)
		triggerType      = getOrEmpty[string]("triggerType", j)
		triggerReference = getOrEmpty[string]("triggerReference", j)

		fillsAny          = getOrEmpty[[]any]("fills", j)
		filledAmount      = getOrEmpty[string]("filledAmount", j)
		filledAmountQuote = getOrEmpty[string]("filledAmountQuote", j)
		feeCurrency       = getOrEmpty[string]("feeCurrency", j)
		feePaid           = getOrEmpty[string]("feePaid", j)
	)

	if len(fillsAny) > 0 {
		fillsBytes, err := json.Marshal(fillsAny)
		if err != nil {
			return err
		}
		fills := make([]Fill, len(fillsAny))
		if err := json.Unmarshal(fillsBytes, &fills); err != nil {
			return err
		}
		o.Fills = fills
	}

	o.OrderId = orderId
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
	o.FilledAmount = util.IfOrElse(len(filledAmount) > 0, func() float64 { return util.MustFloat64(filledAmount) }, 0)
	o.FilledAmountQuote = util.IfOrElse(len(filledAmountQuote) > 0, func() float64 { return util.MustFloat64(filledAmountQuote) }, 0)
	o.FeeCurrency = feeCurrency
	o.FeePaid = util.IfOrElse(len(feePaid) > 0, func() float64 { return util.MustFloat64(feePaid) }, 0)

	return nil
}
