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

type WsHandler[T any] interface {
	// Subscribe to market
	Subscribe(market string) (<-chan T, error)

	// Unsubscribe from market
	Unsubscribe(market string) error

	// Unsubscribe from every market
	UnsubscribeAll() error
}

type WebSocket interface {
	// Close everything, including subscriptions, underlying websockets, gracefull shutdown...
	Close() error

	// Candles websocket handler to handle candle events and subscriptions
	Candles() CandlesWsHandler

	// Ticker websocket handler to handle ticker events and subscriptions
	Ticker() WsHandler[TickerEvent]

	// Ticker24h websocket handler to handle ticker24h events and subscriptions
	Ticker24h() WsHandler[Ticker24hEvent]

	// Trades websocket handler to handle trade events and subscriptions
	Trades() WsHandler[TradesEvent]

	// Book websocket handler to handle book events and subscriptions
	Book() WsHandler[BookEvent]

	// Account websocket handler to handle account events and subscriptions, requires authentication
	Account(apiKey string, apiSecret string) AccountWsHandler
}

type webSocket struct {
	reconnectCount int64
	autoReconnect  bool
	conn           *websocket.Conn
	writechn       chan WebSocketMessage
	debug          bool

	// websocket handlers
	candleWsHandler    *candleWsHandler
	tickerWsHandler    *tickerWsHandler
	ticker24hWsHandler *ticker24hWsHandler
	tradesWsHandler    *tradesWsHandler
	bookWsHandler      *bookWsHandler
	accountWsHandler   *accountWsHandler
}

func NewWebSocket(options ...Option) (WebSocket, error) {
	conn, err := newConn()
	if err != nil {
		return nil, err
	}

	ws := &webSocket{
		conn:          conn,
		autoReconnect: true,
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

func (ws *webSocket) Candles() CandlesWsHandler {
	ws.candleWsHandler = newCandleWsHandler(ws.writechn)
	return ws.candleWsHandler
}

func (ws *webSocket) Ticker() WsHandler[TickerEvent] {
	ws.tickerWsHandler = newTickerWsHandler(ws.writechn)
	return ws.tickerWsHandler
}

func (ws *webSocket) Ticker24h() WsHandler[Ticker24hEvent] {
	ws.ticker24hWsHandler = newTicker24hWsHandler(ws.writechn)
	return ws.ticker24hWsHandler
}

func (ws *webSocket) Trades() WsHandler[TradesEvent] {
	ws.tradesWsHandler = newTradesWsHandler(ws.writechn)
	return ws.tradesWsHandler
}

func (ws *webSocket) Book() WsHandler[BookEvent] {
	ws.bookWsHandler = newBookWsHandler(ws.writechn)
	return ws.bookWsHandler
}

func (ws *webSocket) Account(apiKey string, apiSecret string) AccountWsHandler {
	ws.accountWsHandler = newAccountWsHandler(apiKey, apiSecret, ws.writechn)
	return ws.accountWsHandler
}

func (ws *webSocket) Close() error {
	defer close(ws.writechn)

	if ws.hasCandleWsHandler() {
		ws.candleWsHandler.UnsubscribeAll()
	}
	if ws.hasTickerWsHandler() {
		ws.tickerWsHandler.UnsubscribeAll()
	}
	if ws.hasTicker24hWsHandler() {
		ws.ticker24hWsHandler.UnsubscribeAll()
	}
	if ws.hasTradesWsHandler() {
		ws.tradesWsHandler.UnsubscribeAll()
	}
	if ws.hasBookWsHandler() {
		ws.bookWsHandler.UnsubscribeAll()
	}
	if ws.hasAccountWsHandler() {
		ws.accountWsHandler.UnsubscribeAll()
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
	log.Logger().Info("Connected...")

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
		log.Logger().Info("Auto reconnect disabled, not reconnecting...")
		return
	}

	log.Logger().Info("Reconnecting...")

	conn, err := newConn()
	if err != nil {
		defer ws.reconnect()

		ws.reconnectCount += 1
		log.Logger().Info("Reconnect failed, retrying in 1 second", "count", ws.reconnectCount)
		time.Sleep(time.Second)
		return
	}
	ws.reconnectCount = 0
	ws.conn = conn

	go ws.readLoop()

	if ws.hasCandleWsHandler() {
		ws.candleWsHandler.reconnect()
	}
	if ws.hasTickerWsHandler() {
		ws.tickerWsHandler.reconnect()
	}
	if ws.hasTicker24hWsHandler() {
		ws.ticker24hWsHandler.reconnect()
	}
	if ws.hasTradesWsHandler() {
		ws.tradesWsHandler.reconnect()
	}
	if ws.hasBookWsHandler() {
		ws.bookWsHandler.reconnect()
	}
	if ws.hasAccountWsHandler() {
		ws.accountWsHandler.reconnect()
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
	if ws.debug {
		log.Logger().Debug("Handling incoming message", "message", string(bytes))
	}

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
	if ws.debug {
		log.Logger().Debug("Handling incoming error", "err", err)
	}

	switch err.Action {
	case ActionAuthenticate.Value:
		log.Logger().Error("Failed to authenticate, wrong apiKey and/or apiSecret")
	default:
		log.Logger().Error("Could not handle error", "action", err.Action, "code", err.Code, "message", err.Message)
	}
}

func (ws *webSocket) handleEvent(e *BaseEvent, bytes []byte) {
	if ws.debug {
		log.Logger().Debug("Handling incoming message", "event", e.Event, "message", string(bytes))
	}

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
	if ws.debug {
		log.Logger().Debug("Received subscribed event")
	}
}

func (ws *webSocket) handleUnsubscribedEvent(bytes []byte) {
	if ws.debug {
		log.Logger().Debug("Received unsubscribed event")
	}
}

func (ws *webSocket) handleCandleEvent(bytes []byte) {
	if ws.debug {
		log.Logger().Debug("Received candles event")
	}
	if ws.hasCandleWsHandler() {
		ws.candleWsHandler.handleMessage(bytes)
	}
}

func (ws *webSocket) handleTickerEvent(bytes []byte) {
	if ws.debug {
		log.Logger().Debug("Received ticker event")
	}
	if ws.hasTickerWsHandler() {
		ws.tickerWsHandler.handleMessage(bytes)
	}
}

func (ws *webSocket) handleTicker24hEvent(bytes []byte) {
	if ws.debug {
		log.Logger().Debug("Received ticker24h event")
	}
	if ws.hasTicker24hWsHandler() {
		ws.ticker24hWsHandler.handleMessage(bytes)
	}
}

func (ws *webSocket) handleTradesEvent(bytes []byte) {
	if ws.debug {
		log.Logger().Debug("Received trades event")
	}
	if ws.hasTradesWsHandler() {
		ws.tradesWsHandler.handleMessage(bytes)
	}
}

func (ws *webSocket) handleBookEvent(bytes []byte) {
	if ws.debug {
		log.Logger().Debug("Received book event")
	}
	if ws.hasBookWsHandler() {
		ws.bookWsHandler.handleMessage(bytes)
	}
}

func (ws *webSocket) handleOrderEvent(bytes []byte) {
	if ws.debug {
		log.Logger().Debug("Received order event")
	}
	if ws.hasAccountWsHandler() {
		ws.accountWsHandler.handleOrderMessage(bytes)
	}
}

func (ws *webSocket) handleFillEvent(bytes []byte) {
	if ws.debug {
		log.Logger().Debug("Received fill event")
	}
	if ws.hasAccountWsHandler() {
		ws.accountWsHandler.handleFillMessage(bytes)
	}
}

func (ws *webSocket) handleAuthEvent(bytes []byte) {
	if ws.debug {
		log.Logger().Debug("Received auth event")
	}
	if ws.hasAccountWsHandler() {
		ws.accountWsHandler.handleAuthMessage(bytes)
	}
}

func (ws *webSocket) hasCandleWsHandler() bool {
	return ws.candleWsHandler != nil
}

func (ws *webSocket) hasTickerWsHandler() bool {
	return ws.tickerWsHandler != nil
}

func (ws *webSocket) hasTicker24hWsHandler() bool {
	return ws.ticker24hWsHandler != nil
}

func (ws *webSocket) hasTradesWsHandler() bool {
	return ws.tradesWsHandler != nil
}

func (ws *webSocket) hasBookWsHandler() bool {
	return ws.bookWsHandler != nil
}

func (ws *webSocket) hasAccountWsHandler() bool {
	return ws.accountWsHandler != nil
}
