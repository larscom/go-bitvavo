package jsond

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
