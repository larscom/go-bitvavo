package ws

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/larscom/go-bitvavo/v2/types"
	"github.com/rs/zerolog/log"

	"github.com/goccy/go-json"
	"github.com/gorilla/websocket"
)

const (
	wsUrl            = "wss://ws.bitvavo.com/v2"
	readLimit        = 655350
	handshakeTimeout = 45 * time.Second
	defaultBuffSize  = 50
)

var (
	errNoSubscriptionActive      = func(market string) error { return fmt.Errorf("no active subscription for market: %s", market) }
	errSubscriptionAlreadyActive = func(market string) error { return fmt.Errorf("subscription already active for market: %s", market) }
	errAuthenticationFailed      = errors.New("could not subscribe, authentication failed")
	errEventHandler              = errors.New("could not handle event")
)

type EventHandler[T any] interface {
	// Subscribe to markets.
	// You can set the buffSize for the channel.
	//
	// If you have many subscriptions at once you may need to increase the buffSize
	//
	// Default buffSize: 50
	Subscribe(markets []string, buffSize ...uint64) (<-chan T, error)

	// Unsubscribe from markets.
	Unsubscribe(markets []string) error

	// Unsubscribe from every market.
	UnsubscribeAll() error
}

type WsClient interface {
	// Close everything, including subscriptions, underlying websockets, graceful shutdown...
	Close() error

	// Candles event handler to handle candle events and subscriptions.
	Candles() CandlesEventHandler

	// Ticker event handler to handle ticker events and subscriptions.
	Ticker() EventHandler[TickerEvent]

	// Ticker24h event handler to handle ticker24h events and subscriptions.
	Ticker24h() EventHandler[Ticker24hEvent]

	// Trades event handler to handle trade events and subscriptions.
	Trades() EventHandler[TradesEvent]

	// Book event handler to handle book events and subscriptions.
	Book() EventHandler[BookEvent]

	// Account event handler to handle order/fill events, requires authentication.
	Account(apiKey string, apiSecret string) AccountEventHandler
}

type handler interface {
	UnsubscribeAll() error

	reconnect()

	handleMessage(e WsEvent, bytes []byte)
}

type wsClient struct {
	reconnectCount uint64
	autoReconnect  bool
	conn           *websocket.Conn
	writechn       chan WebSocketMessage
	errchn         chan<- error

	// all registered event handlers
	handlers []handler
}

func NewWsClient(options ...Option) (WsClient, error) {
	conn, err := newConn()
	if err != nil {
		return nil, err
	}

	ws := &wsClient{
		conn:          conn,
		autoReconnect: true,
		writechn:      make(chan WebSocketMessage),
		handlers:      make([]handler, 0),
	}
	for _, opt := range options {
		opt(ws)
	}

	go ws.writeLoop()
	go ws.readLoop()

	return ws, nil
}

type Option func(*wsClient)

// Receive websocket connection errors (e.g. reconnect error, auth error, write failed, read failed)
func WithErrorChannel(errchn chan<- error) Option {
	return func(ws *wsClient) {
		ws.errchn = errchn
	}
}

// Auto reconnect if websocket disconnects.
// default: true
func WithAutoReconnect(autoReconnect bool) Option {
	return func(ws *wsClient) {
		ws.autoReconnect = autoReconnect
	}
}

// The buff size for the write channel, by default the write channel is unbuffered.
// The write channel writes messages to the websocket.
func WithWriteBuffSize(buffSize uint64) Option {
	return func(ws *wsClient) {
		ws.writechn = make(chan WebSocketMessage, buffSize)
	}
}

func (ws *wsClient) Candles() CandlesEventHandler {
	for _, h := range ws.handlers {
		if handler, ok := h.(*candlesEventHandler); ok {
			return handler
		}
	}

	handler := newCandlesEventHandler(ws.writechn)
	ws.handlers = append(ws.handlers, handler)

	return handler
}

func (ws *wsClient) Ticker() EventHandler[TickerEvent] {
	for _, h := range ws.handlers {
		if handler, ok := h.(*tickerEventHandler); ok {
			return handler
		}
	}

	handler := newTickerEventHandler(ws.writechn)
	ws.handlers = append(ws.handlers, handler)

	return handler
}

func (ws *wsClient) Ticker24h() EventHandler[Ticker24hEvent] {
	for _, h := range ws.handlers {
		if handler, ok := h.(*ticker24hEventHandler); ok {
			return handler
		}
	}

	handler := newTicker24hEventHandler(ws.writechn)
	ws.handlers = append(ws.handlers, handler)

	return handler
}

func (ws *wsClient) Trades() EventHandler[TradesEvent] {
	for _, h := range ws.handlers {
		if handler, ok := h.(*tradesEventHandler); ok {
			return handler
		}
	}

	handler := newTradesEventHandler(ws.writechn)
	ws.handlers = append(ws.handlers, handler)

	return handler
}

func (ws *wsClient) Book() EventHandler[BookEvent] {
	for _, h := range ws.handlers {
		if handler, ok := h.(*bookEventHandler); ok {
			return handler
		}
	}

	handler := newBookEventHandler(ws.writechn)
	ws.handlers = append(ws.handlers, handler)

	return handler
}

func (ws *wsClient) Account(apiKey string, apiSecret string) AccountEventHandler {
	for _, h := range ws.handlers {
		if handler, ok := h.(*accountEventHandler); ok {
			return handler
		}
	}

	handler := newAccountEventHandler(apiKey, apiSecret, ws.writechn)
	ws.handlers = append(ws.handlers, handler)

	return handler
}

func (ws *wsClient) Close() error {
	defer close(ws.writechn)

	for _, handler := range ws.handlers {
		handler.UnsubscribeAll()
	}

	if ws.hasErrorChannel() {
		close(ws.errchn)
	}

	return ws.conn.Close()
}

func newConn() (*websocket.Conn, error) {
	dialer := websocket.Dialer{
		Proxy:             http.ProxyFromEnvironment,
		HandshakeTimeout:  handshakeTimeout,
		EnableCompression: false,
	}

	conn, _, err := dialer.Dial(wsUrl, nil)
	if err != nil {
		return nil, err
	}
	conn.SetReadLimit(readLimit)

	return conn, nil
}

func (ws *wsClient) writeLoop() {
	for msg := range ws.writechn {
		if err := ws.conn.WriteJSON(msg); err != nil {
			log.Err(err).Msg("Write failed")
			if ws.hasErrorChannel() {
				ws.errchn <- err
			}
		}
	}
}

func (ws *wsClient) readLoop() {
	log.Debug().Msg("Connected...")

	for {
		_, bytes, err := ws.conn.ReadMessage()
		if err != nil {
			defer ws.reconnect()

			log.Err(err).Msg("Read failed")
			if ws.hasErrorChannel() {
				ws.errchn <- err
			}

			return
		}
		ws.handleMessage(bytes)
	}
}

func (ws *wsClient) reconnect() {
	if !ws.autoReconnect {
		log.Debug().Msg("Auto reconnect disabled, not reconnecting...")
		return
	}

	log.Debug().Msg("Reconnecting...")

	conn, err := newConn()
	if err != nil {
		defer ws.reconnect()

		ws.reconnectCount += 1
		log.Error().
			Uint64("count", ws.reconnectCount).
			Msg("Reconnect failed, retrying in 1 second")

		if ws.hasErrorChannel() {
			ws.errchn <- err
		}
		time.Sleep(time.Second)
		return
	}
	ws.reconnectCount = 0
	ws.conn = conn

	go ws.readLoop()

	for _, handler := range ws.handlers {
		handler.reconnect()
	}
}

func newWebSocketMessage(action Action, channelName ChannelName, markets []string) WebSocketMessage {
	return WebSocketMessage{
		Action: action.Value,
		Channels: []Channel{
			{
				Name:    channelName.Value,
				Markets: markets,
			},
		},
	}
}

func (ws *wsClient) handleMessage(bytes []byte) {
	log.Debug().Str("message", string(bytes)).Msg("Handling incoming message")

	var baseEvent *BaseEvent
	if err := json.Unmarshal(bytes, &baseEvent); err != nil {
		var wsError *types.BitvavoErr
		if err := json.Unmarshal(bytes, &wsError); err != nil {
			log.Err(err).Str("message", string(bytes)).Msg("Don't know how to handle this message")
		} else {
			ws.handlError(wsError)
		}
	} else {
		ws.handleEvent(baseEvent, bytes)
	}
}

func (ws *wsClient) handlError(err *types.BitvavoErr) {
	log.Debug().Str("error", err.Error()).Msg("Handling incoming error")

	switch err.Action {
	case actionAuthenticate.Value:
		log.Err(err).Msg("Failed to authenticate, wrong apiKey and/or apiSecret")
	default:
		log.Err(err).Msg("Could not handle error")
	}

	if ws.hasErrorChannel() {
		ws.errchn <- err
	}
}

func (ws *wsClient) handleEvent(e *BaseEvent, bytes []byte) {
	log.Debug().Str("event", e.Event).Msg("Handling incoming message")

	switch e.Event {
	// public
	case wsEventSubscribed.Value:
		ws.handleSubscribedEvent(bytes)
	case wsEventUnsubscribed.Value:
		ws.handleUnsubscribedEvent(bytes)
	case wsEventCandles.Value:
		ws.handleCandleEvent(wsEventCandles, bytes)
	case wsEventTicker.Value:
		ws.handleTickerEvent(wsEventTicker, bytes)
	case wsEventTicker24h.Value:
		ws.handleTicker24hEvent(wsEventTicker24h, bytes)
	case wsEventTrades.Value:
		ws.handleTradesEvent(wsEventTrades, bytes)
	case wsEventBook.Value:
		ws.handleBookEvent(wsEventBook, bytes)

	// authenticated
	case wsEventAuth.Value:
		ws.handleAuthEvent(wsEventAuth, bytes)
	case wsEventOrder.Value:
		ws.handleOrderEvent(wsEventOrder, bytes)
	case wsEventFill.Value:
		ws.handleFillEvent(wsEventFill, bytes)

	default:
		log.Error().Msg("Could not handle event, invalid parameters provided?")
		if ws.hasErrorChannel() {
			ws.errchn <- errEventHandler
		}
	}
}

func (ws *wsClient) handleSubscribedEvent(bytes []byte) {
	log.Debug().Str("message", string(bytes)).Msg("Received subscribed event")
}

func (ws *wsClient) handleUnsubscribedEvent(bytes []byte) {
	log.Debug().Str("message", string(bytes)).Msg("Received unsubscribed event")
}

func (ws *wsClient) handleCandleEvent(e WsEvent, bytes []byte) {
	log.Debug().Str("message", string(bytes)).Msg("Received candles event")

	for _, h := range ws.handlers {
		if handler, ok := h.(*candlesEventHandler); ok {
			handler.handleMessage(e, bytes)
		}
	}
}

func (ws *wsClient) handleTickerEvent(e WsEvent, bytes []byte) {
	log.Debug().Str("message", string(bytes)).Msg("Received ticker event")

	for _, h := range ws.handlers {
		if handler, ok := h.(*tickerEventHandler); ok {
			handler.handleMessage(e, bytes)
		}
	}
}

func (ws *wsClient) handleTicker24hEvent(e WsEvent, bytes []byte) {
	log.Debug().Str("message", string(bytes)).Msg("Received ticker24h event")

	for _, h := range ws.handlers {
		if handler, ok := h.(*ticker24hEventHandler); ok {
			handler.handleMessage(e, bytes)
		}
	}
}

func (ws *wsClient) handleTradesEvent(e WsEvent, bytes []byte) {
	log.Debug().Str("message", string(bytes)).Msg("Received trades event")

	for _, h := range ws.handlers {
		if handler, ok := h.(*tradesEventHandler); ok {
			handler.handleMessage(e, bytes)
		}
	}
}

func (ws *wsClient) handleBookEvent(e WsEvent, bytes []byte) {
	log.Debug().Str("message", string(bytes)).Msg("Received book event")

	for _, h := range ws.handlers {
		if handler, ok := h.(*bookEventHandler); ok {
			handler.handleMessage(e, bytes)
		}
	}
}

func (ws *wsClient) handleOrderEvent(e WsEvent, bytes []byte) {
	log.Debug().Str("message", string(bytes)).Msg("Received order event")

	for _, h := range ws.handlers {
		if handler, ok := h.(*accountEventHandler); ok {
			handler.handleMessage(e, bytes)
		}
	}
}

func (ws *wsClient) handleFillEvent(e WsEvent, bytes []byte) {
	log.Debug().Str("message", string(bytes)).Msg("Received fill event")

	for _, h := range ws.handlers {
		if handler, ok := h.(*accountEventHandler); ok {
			handler.handleMessage(e, bytes)
		}
	}
}

func (ws *wsClient) handleAuthEvent(e WsEvent, bytes []byte) {
	log.Debug().Str("message", string(bytes)).Msg("Received auth event")

	for _, h := range ws.handlers {
		if handler, ok := h.(*accountEventHandler); ok {
			handler.handleMessage(e, bytes)
		}
	}
}

func (ws *wsClient) hasErrorChannel() bool {
	return ws.errchn != nil
}
