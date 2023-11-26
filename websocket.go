package bitvavo

import (
	"net/http"
	"time"

	"github.com/larscom/go-bitvavo/v2/log"

	"github.com/goccy/go-json"
	"github.com/gorilla/websocket"
)

const (
	wsUrl            = "wss://ws.bitvavo.com/v2"
	readLimit        = 655350
	handshakeTimeout = 45 * time.Second
)

type EventHandler[T any] interface {
	// Subscribe to market.
	// You can set the buffSize for the underlying channel, 0 for no buffer.
	Subscribe(market string, buffSize uint64) (<-chan T, error)

	// Unsubscribe from market.
	Unsubscribe(market string) error

	// Unsubscribe from every market.
	UnsubscribeAll() error
}

type WebSocket interface {
	// Close everything, including subscriptions, underlying websockets, gracefull shutdown...
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

	// Account event handler to handle account subscription and order/fill events, requires authentication.
	Account(apiKey string, apiSecret string) AccountEventHandler
}

type webSocket struct {
	reconnectCount uint64
	autoReconnect  bool
	conn           *websocket.Conn
	writechn       chan WebSocketMessage
	debug          bool

	// public
	candlesEventHandler   *candlesEventHandler
	tickerEventHandler    *tickerEventHandler
	ticker24hEventHandler *ticker24hEventHandler
	tradesEventHandler    *tradesEventHandler
	bookEventHandler      *bookEventHandler

	// authenticated
	accountEventHandler *accountEventHandler
	windowTimeMs        uint64
}

func NewWebSocket(options ...Option) (WebSocket, error) {
	conn, err := newConn()
	if err != nil {
		return nil, err
	}

	ws := &webSocket{
		conn:          conn,
		autoReconnect: true,
		windowTimeMs:  10000,
		writechn:      make(chan WebSocketMessage),
	}

	for _, opt := range options {
		opt(ws)
	}

	go ws.writeLoop()
	go ws.readLoop()

	return ws, nil
}

type Option func(*webSocket)

// Enable debug logging.
// default: false
func WithDebug(debug bool) Option {
	return func(ws *webSocket) {
		ws.debug = debug
	}
}

// Auto reconnect if websocket disconnects.
// default: true
func WithAutoReconnect(autoReconnect bool) Option {
	return func(ws *webSocket) {
		ws.autoReconnect = autoReconnect
	}
}

// The time in milliseconds that your request is allowed to execute in.
// The default value is 10000 (10s), the maximum value is 60000 (60s).
func WithWindowTime(windowTimeMs uint64) Option {
	return func(ws *webSocket) {
		if windowTimeMs > 60000 {
			windowTimeMs = 60000
		}
		ws.windowTimeMs = windowTimeMs
	}
}

// The buff size for the write channel, by default the write channel is unbuffered.
// The write channel writes messages to the websocket.
func WithWriteBuffSize(buffSize uint64) Option {
	return func(ws *webSocket) {
		ws.writechn = make(chan WebSocketMessage, buffSize)
	}
}

func (ws *webSocket) Candles() CandlesEventHandler {
	ws.candlesEventHandler = newCandlesEventHandler(ws.writechn)
	return ws.candlesEventHandler
}

func (ws *webSocket) Ticker() EventHandler[TickerEvent] {
	ws.tickerEventHandler = newTickerEventHandler(ws.writechn)
	return ws.tickerEventHandler
}

func (ws *webSocket) Ticker24h() EventHandler[Ticker24hEvent] {
	ws.ticker24hEventHandler = newTicker24hEventHandler(ws.writechn)
	return ws.ticker24hEventHandler
}

func (ws *webSocket) Trades() EventHandler[TradesEvent] {
	ws.tradesEventHandler = newTradesEventHandler(ws.writechn)
	return ws.tradesEventHandler
}

func (ws *webSocket) Book() EventHandler[BookEvent] {
	ws.bookEventHandler = newBookEventHandler(ws.writechn)
	return ws.bookEventHandler
}

func (ws *webSocket) Account(apiKey string, apiSecret string) AccountEventHandler {
	ws.accountEventHandler = newAccountEventHandler(apiKey, apiSecret, ws.windowTimeMs, ws.writechn)
	return ws.accountEventHandler
}

func (ws *webSocket) Close() error {
	defer close(ws.writechn)

	if ws.hasCandleWsHandler() {
		ws.candlesEventHandler.UnsubscribeAll()
	}
	if ws.hasTickerWsHandler() {
		ws.tickerEventHandler.UnsubscribeAll()
	}
	if ws.hasTicker24hWsHandler() {
		ws.ticker24hEventHandler.UnsubscribeAll()
	}
	if ws.hasTradesWsHandler() {
		ws.tradesEventHandler.UnsubscribeAll()
	}
	if ws.hasBookWsHandler() {
		ws.bookEventHandler.UnsubscribeAll()
	}
	if ws.hasAccountWsHandler() {
		ws.accountEventHandler.UnsubscribeAll()
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

func (ws *webSocket) writeLoop() {
	for msg := range ws.writechn {
		if err := ws.conn.WriteJSON(msg); err != nil {
			log.Logger().Error("Write failed", "error", err.Error())
		}
	}
}

func (ws *webSocket) readLoop() {
	ws.logDebug("Connected...")

	for {
		_, bytes, err := ws.conn.ReadMessage()
		if err != nil {
			defer ws.reconnect()
			return
		}
		ws.handleMessage(bytes)
	}
}

func (ws *webSocket) reconnect() {
	if !ws.autoReconnect {
		ws.logDebug("Auto reconnect disabled, not reconnecting...")
		return
	}

	ws.logDebug("Reconnecting...")

	conn, err := newConn()
	if err != nil {
		defer ws.reconnect()

		ws.reconnectCount += 1
		log.Logger().Error("Reconnect failed, retrying in 1 second", "count", ws.reconnectCount)
		time.Sleep(time.Second)
		return
	}
	ws.reconnectCount = 0
	ws.conn = conn

	go ws.readLoop()

	if ws.hasCandleWsHandler() {
		ws.candlesEventHandler.reconnect()
	}
	if ws.hasTickerWsHandler() {
		ws.tickerEventHandler.reconnect()
	}
	if ws.hasTicker24hWsHandler() {
		ws.ticker24hEventHandler.reconnect()
	}
	if ws.hasTradesWsHandler() {
		ws.tradesEventHandler.reconnect()
	}
	if ws.hasBookWsHandler() {
		ws.bookEventHandler.reconnect()
	}
	if ws.hasAccountWsHandler() {
		ws.accountEventHandler.reconnect()
	}
}

func newWebSocketMessage(action Action, channelName ChannelName, market string) WebSocketMessage {
	return WebSocketMessage{
		Action: action.Value,
		Channels: []Channel{
			{
				Name:    channelName.Value,
				Markets: []string{market},
			},
		},
	}
}

func (ws *webSocket) handleMessage(bytes []byte) {
	ws.logDebug("Handling incoming message", "message", string(bytes))

	var baseEvent *BaseEvent
	if err := json.Unmarshal(bytes, &baseEvent); err != nil {
		var wsError *WebSocketErr
		if err := json.Unmarshal(bytes, &wsError); err != nil {
			log.Logger().Error("Don't know how to handle this message", "message", string(bytes))
		} else {
			ws.handlError(wsError)
		}
	} else {
		ws.handleEvent(baseEvent, bytes)
	}
}

func (ws *webSocket) handlError(err *WebSocketErr) {
	ws.logDebug("Handling incoming error", "err", err)

	switch err.Action {
	case ActionAuthenticate.Value:
		log.Logger().Error("Failed to authenticate, wrong apiKey and/or apiSecret")
	default:
		log.Logger().Error("Could not handle error", "action", err.Action, "code", err.Code, "message", err.Message)
	}
}

func (ws *webSocket) handleEvent(e *BaseEvent, bytes []byte) {
	ws.logDebug("Handling incoming message", "event", e.Event, "message", string(bytes))

	switch e.Event {
	// public
	case WsEventSubscribed.Value:
		ws.handleSubscribedEvent(bytes)
	case WsEventUnsubscribed.Value:
		ws.handleUnsubscribedEvent(bytes)
	case WsEventCandles.Value:
		ws.handleCandleEvent(bytes)
	case WsEventTicker.Value:
		ws.handleTickerEvent(bytes)
	case WsEventTicker24h.Value:
		ws.handleTicker24hEvent(bytes)
	case WsEventTrades.Value:
		ws.handleTradesEvent(bytes)
	case WsEventBook.Value:
		ws.handleBookEvent(bytes)

	// authenticated
	case WsEventAuth.Value:
		ws.handleAuthEvent(bytes)
	case WsEventOrder.Value:
		ws.handleOrderEvent(bytes)
	case WsEventFill.Value:
		ws.handleFillEvent(bytes)

	default:
		log.Logger().Error("Could not handle event, invalid parameters provided?")
	}
}

func (ws *webSocket) handleSubscribedEvent(bytes []byte) {
	ws.logDebug("Received subscribed event")
}

func (ws *webSocket) handleUnsubscribedEvent(bytes []byte) {
	ws.logDebug("Received unsubscribed event")
}

func (ws *webSocket) handleCandleEvent(bytes []byte) {
	ws.logDebug("Received candles event")

	if ws.hasCandleWsHandler() {
		ws.candlesEventHandler.handleMessage(bytes)
	}
}

func (ws *webSocket) handleTickerEvent(bytes []byte) {
	ws.logDebug("Received ticker event")

	if ws.hasTickerWsHandler() {
		ws.tickerEventHandler.handleMessage(bytes)
	}
}

func (ws *webSocket) handleTicker24hEvent(bytes []byte) {
	ws.logDebug("Received ticker24h event")

	if ws.hasTicker24hWsHandler() {
		ws.ticker24hEventHandler.handleMessage(bytes)
	}
}

func (ws *webSocket) handleTradesEvent(bytes []byte) {
	ws.logDebug("Received trades event")

	if ws.hasTradesWsHandler() {
		ws.tradesEventHandler.handleMessage(bytes)
	}
}

func (ws *webSocket) handleBookEvent(bytes []byte) {
	ws.logDebug("Received book event")

	if ws.hasBookWsHandler() {
		ws.bookEventHandler.handleMessage(bytes)
	}
}

func (ws *webSocket) handleOrderEvent(bytes []byte) {
	ws.logDebug("Received order event")

	if ws.hasAccountWsHandler() {
		ws.accountEventHandler.handleOrderMessage(bytes)
	}
}

func (ws *webSocket) handleFillEvent(bytes []byte) {
	ws.logDebug("Received fill event")

	if ws.hasAccountWsHandler() {
		ws.accountEventHandler.handleFillMessage(bytes)
	}
}

func (ws *webSocket) handleAuthEvent(bytes []byte) {
	ws.logDebug("Received auth event")

	if ws.hasAccountWsHandler() {
		ws.accountEventHandler.handleAuthMessage(bytes)
	}
}

func (ws *webSocket) hasCandleWsHandler() bool {
	return ws.candlesEventHandler != nil
}

func (ws *webSocket) hasTickerWsHandler() bool {
	return ws.tickerEventHandler != nil
}

func (ws *webSocket) hasTicker24hWsHandler() bool {
	return ws.ticker24hEventHandler != nil
}

func (ws *webSocket) hasTradesWsHandler() bool {
	return ws.tradesEventHandler != nil
}

func (ws *webSocket) hasBookWsHandler() bool {
	return ws.bookEventHandler != nil
}

func (ws *webSocket) hasAccountWsHandler() bool {
	return ws.accountEventHandler != nil
}

func (ws *webSocket) logDebug(message string, args ...any) {
	if ws.debug {
		log.Logger().Debug(message, args...)
	}
}
