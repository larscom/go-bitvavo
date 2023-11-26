package bitvavo

import "github.com/orsinium-labs/enum"

const ZERO = float64(0)

type WsEvent enum.Member[string]

var (
	WsEventSubscribed   = WsEvent{"subscribed"}
	WsEventUnsubscribed = WsEvent{"unsubscribed"}
	WsEventCandles      = WsEvent{"candle"}
	WsEventTicker       = WsEvent{"ticker"}
	WsEventTicker24h    = WsEvent{"ticker24h"}
	WsEventTrades       = WsEvent{"trade"}
	WsEventBook         = WsEvent{"book"}
	WsEventAuth         = WsEvent{"authenticate"}
	WsEventAccount      = WsEvent{"account"}
	WsEventOrder        = WsEvent{"order"}
	WsEventFill         = WsEvent{"fill"}
)

type Action enum.Member[string]

var (
	ActionSubscribe    = Action{"subscribe"}
	ActionUnsubscribe  = Action{"unsubscribe"}
	ActionAuthenticate = Action{"authenticate"}
)

type ChannelName enum.Member[string]

var (
	ChannelNameCandles   = ChannelName{"candles"}
	ChannelNameTicker    = ChannelName{"ticker"}
	ChannelNameTicker24h = ChannelName{"ticker24h"}
	ChannelNameTrades    = ChannelName{"trades"}
	ChannelNameBook      = ChannelName{"book"}
	ChannelNameAccount   = ChannelName{"account"}
)

type SubscribedEvent struct {
	// Describes the returned event over the socket.
	Event string `json:"event"`

	// Subscriptions map[event][]markets
	Subscriptions map[string][]string `json:"subscriptions"`
}

type AuthEvent struct {
	// Describes the returned event over the socket.
	Event string `json:"event"`

	// Whether the user is authenticated.
	Authenticated bool `json:"authenticated"`
}

type BaseEvent struct {
	Event string `json:"event"`
}

type WebSocketErr struct {
	Action  string `json:"action"`
	Code    int    `json:"errorCode"`
	Message string `json:"error"`
}

type WebSocketMessage struct {
	Action   string    `json:"action"`
	Channels []Channel `json:"channels,omitempty"`

	// Api Key.
	Key string `json:"key,omitempty"`
	// SHA256 HMAC hex digest of timestamp + method + url + body.
	Signature string `json:"signature,omitempty"`
	// The current timestamp in milliseconds since 1 Jan 1970.
	Timestamp int64 `json:"timestamp,omitempty"`
	// The window that allows execution of your request in milliseconds since 1 Jan 1970. The default value is 10000 (10s) and maximum value is 60000 (60s).
	Window uint64 `json:"window,omitempty"`
}

type Channel struct {
	Name      string   `json:"name"`
	Intervals []string `json:"interval,omitempty"`
	Markets   []string `json:"markets,omitempty"`
}
