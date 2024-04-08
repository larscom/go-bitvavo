package ws

import (
	"fmt"

	"github.com/goccy/go-json"
)

type AuthEvent struct {
	// Describes the returned event over the socket.
	Event string `json:"event"`

	// Whether the user is authenticated.
	Authenticated bool `json:"authenticated"`
}

type BaseEvent struct {
	Event WsEvent `json:"event"`
}

func (b *BaseEvent) UnmarshalJSON(bytes []byte) error {
	var j map[string]any

	if err := json.Unmarshal(bytes, &j); err != nil {
		return err
	}

	e := j["event"].(string)

	event := wsEvents.Parse(e)
	if event == nil {
		return fmt.Errorf("unknown event type: %s", e)
	}

	b.Event = *event

	return nil
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
