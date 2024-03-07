package ws

type AuthEvent struct {
	// Describes the returned event over the socket.
	Event string `json:"event"`

	// Whether the user is authenticated.
	Authenticated bool `json:"authenticated"`
}

type BaseEvent struct {
	Event string `json:"event"`
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
}

type Channel struct {
	Name      string   `json:"name"`
	Intervals []string `json:"interval,omitempty"`
	Markets   []string `json:"markets,omitempty"`
}
