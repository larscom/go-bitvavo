package wsc

import (
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/larscom/go-bitvavo/v2/types"

	"github.com/goccy/go-json"
	"github.com/gorilla/websocket"
)

const (
	wsUrl            = "wss://ws.bitvavo.com/v2"
	readLimit        = 655350
	handshakeTimeout = 45 * time.Second
)
const DefaultBuffSize = 50

var (
	ErrNoSubscriptionActive      = errors.New("no subscription active")
	ErrSubscriptionAlreadyActive = errors.New("subscription already active")
	ErrAuthenticationFailed      = errors.New("could not subscribe, authentication failed")
	ErrEventHandler              = errors.New("could not handle event")
)

type EventHandler[T any] interface {
	// Subscribe to market.
	// You can set the buffSize for the channel.
	//
	// If you have many subscriptions at once you may need to increase the buffSize
	//
	// Default buffSize: 50
	Subscribe(market string, buffSize ...uint64) (<-chan T, error)

	// Unsubscribe from market.
	Unsubscribe(market string) error

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

	// Account event handler to handle account subscription and order/fill events, requires authentication.
	Account(apiKey string, apiSecret string) AccountEventHandler
}

type wsClient struct {
	reconnectCount uint64
	autoReconnect  bool
	conn           *websocket.Conn
	writechn       chan WebSocketMessage
	errchn         chan<- error

	// public
	candlesEventHandler   *candlesEventHandler
	tickerEventHandler    *tickerEventHandler
	ticker24hEventHandler *ticker24hEventHandler
	tradesEventHandler    *tradesEventHandler
	bookEventHandler      *bookEventHandler

	// authenticated
	accountEventHandler *accountEventHandler
}

func NewWsClient(options ...Option) (WsClient, error) {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	conn, err := newConn()
	if err != nil {
		return nil, err
	}

	ws := &wsClient{
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

type Option func(*wsClient)

// Enable debug logging
func WithDebug() Option {
	return func(ws *wsClient) {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})))
	}
}

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
	if ws.hasCandleHandler() {
		return ws.candlesEventHandler
	}

	ws.candlesEventHandler = newCandlesEventHandler(ws.writechn)
	return ws.candlesEventHandler
}

func (ws *wsClient) Ticker() EventHandler[TickerEvent] {
	if ws.hasTickerHandler() {
		return ws.tickerEventHandler
	}

	ws.tickerEventHandler = newTickerEventHandler(ws.writechn)
	return ws.tickerEventHandler
}

func (ws *wsClient) Ticker24h() EventHandler[Ticker24hEvent] {
	if ws.hasTicker24hHandler() {
		return ws.ticker24hEventHandler
	}

	ws.ticker24hEventHandler = newTicker24hEventHandler(ws.writechn)
	return ws.ticker24hEventHandler
}

func (ws *wsClient) Trades() EventHandler[TradesEvent] {
	if ws.hasTradesHandler() {
		return ws.tradesEventHandler
	}

	ws.tradesEventHandler = newTradesEventHandler(ws.writechn)
	return ws.tradesEventHandler
}

func (ws *wsClient) Book() EventHandler[BookEvent] {
	if ws.hasBookHandler() {
		return ws.bookEventHandler
	}

	ws.bookEventHandler = newBookEventHandler(ws.writechn)
	return ws.bookEventHandler
}

func (ws *wsClient) Account(apiKey string, apiSecret string) AccountEventHandler {
	if ws.hasAccountHandler() {
		return ws.accountEventHandler
	}

	ws.accountEventHandler = newAccountEventHandler(apiKey, apiSecret, ws.writechn)
	return ws.accountEventHandler
}

func (ws *wsClient) Close() error {
	close(ws.writechn)

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
			slog.Error("Write failed", "error", err.Error())
			if ws.hasErrorChannel() {
				ws.errchn <- err
			}
		}
	}
}

func (ws *wsClient) readLoop() {
	slog.Debug("Connected...")

	for {
		_, bytes, err := ws.conn.ReadMessage()
		if err != nil {
			defer ws.reconnect()
			slog.Error("Read failed", "error", err.Error())
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
		slog.Debug("Auto reconnect disabled, not reconnecting...")
		return
	}

	slog.Debug("Reconnecting...")

	conn, err := newConn()
	if err != nil {
		defer ws.reconnect()

		ws.reconnectCount += 1
		slog.Error("Reconnect failed, retrying in 1 second", "count", ws.reconnectCount)
		if ws.hasErrorChannel() {
			ws.errchn <- err
		}
		time.Sleep(time.Second)
		return
	}
	ws.reconnectCount = 0
	ws.conn = conn

	go ws.readLoop()

	if ws.hasCandleHandler() {
		ws.candlesEventHandler.reconnect()
	}
	if ws.hasTickerHandler() {
		ws.tickerEventHandler.reconnect()
	}
	if ws.hasTicker24hHandler() {
		ws.ticker24hEventHandler.reconnect()
	}
	if ws.hasTradesHandler() {
		ws.tradesEventHandler.reconnect()
	}
	if ws.hasBookHandler() {
		ws.bookEventHandler.reconnect()
	}
	if ws.hasAccountHandler() {
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

func (ws *wsClient) handleMessage(bytes []byte) {
	slog.Debug("Handling incoming message", "message", string(bytes))

	var baseEvent *BaseEvent
	if err := json.Unmarshal(bytes, &baseEvent); err != nil {
		var wsError *types.BitvavoErr
		if err := json.Unmarshal(bytes, &wsError); err != nil {
			slog.Error("Don't know how to handle this message", "message", string(bytes))
		} else {
			ws.handlError(wsError)
		}
	} else {
		ws.handleEvent(baseEvent, bytes)
	}
}

func (ws *wsClient) handlError(err *types.BitvavoErr) {
	slog.Debug("Handling incoming error", "err", err)

	switch err.Action {
	case actionAuthenticate.Value:
		slog.Error("Failed to authenticate, wrong apiKey and/or apiSecret")
	default:
		slog.Error("Could not handle error", "action", err.Action, "code", err.Code, "message", err.Message)
	}

	if ws.hasErrorChannel() {
		ws.errchn <- err
	}
}

func (ws *wsClient) handleEvent(e *BaseEvent, bytes []byte) {
	slog.Debug("Handling incoming message", "event", e.Event, "message", string(bytes))

	switch e.Event {
	// public
	case wsEventSubscribed.Value:
		ws.handleSubscribedEvent(bytes)
	case wsEventUnsubscribed.Value:
		ws.handleUnsubscribedEvent(bytes)
	case wsEventCandles.Value:
		ws.handleCandleEvent(bytes)
	case wsEventTicker.Value:
		ws.handleTickerEvent(bytes)
	case wsEventTicker24h.Value:
		ws.handleTicker24hEvent(bytes)
	case wsEventTrades.Value:
		ws.handleTradesEvent(bytes)
	case wsEventBook.Value:
		ws.handleBookEvent(bytes)

	// authenticated
	case wsEventAuth.Value:
		ws.handleAuthEvent(bytes)
	case wsEventOrder.Value:
		ws.handleOrderEvent(bytes)
	case wsEventFill.Value:
		ws.handleFillEvent(bytes)

	default:
		slog.Error("Could not handle event, invalid parameters provided?")
		if ws.hasErrorChannel() {
			ws.errchn <- ErrEventHandler
		}
	}
}

func (ws *wsClient) handleSubscribedEvent(bytes []byte) {
	slog.Debug("Received subscribed event")
}

func (ws *wsClient) handleUnsubscribedEvent(bytes []byte) {
	slog.Debug("Received unsubscribed event")
}

func (ws *wsClient) handleCandleEvent(bytes []byte) {
	slog.Debug("Received candles event")

	if ws.hasCandleHandler() {
		ws.candlesEventHandler.handleMessage(bytes)
	}
}

func (ws *wsClient) handleTickerEvent(bytes []byte) {
	slog.Debug("Received ticker event")

	if ws.hasTickerHandler() {
		ws.tickerEventHandler.handleMessage(bytes)
	}
}

func (ws *wsClient) handleTicker24hEvent(bytes []byte) {
	slog.Debug("Received ticker24h event")

	if ws.hasTicker24hHandler() {
		ws.ticker24hEventHandler.handleMessage(bytes)
	}
}

func (ws *wsClient) handleTradesEvent(bytes []byte) {
	slog.Debug("Received trades event")

	if ws.hasTradesHandler() {
		ws.tradesEventHandler.handleMessage(bytes)
	}
}

func (ws *wsClient) handleBookEvent(bytes []byte) {
	slog.Debug("Received book event")

	if ws.hasBookHandler() {
		ws.bookEventHandler.handleMessage(bytes)
	}
}

func (ws *wsClient) handleOrderEvent(bytes []byte) {
	slog.Debug("Received order event")

	if ws.hasAccountHandler() {
		ws.accountEventHandler.handleOrderMessage(bytes)
	}
}

func (ws *wsClient) handleFillEvent(bytes []byte) {
	slog.Debug("Received fill event")

	if ws.hasAccountHandler() {
		ws.accountEventHandler.handleFillMessage(bytes)
	}
}

func (ws *wsClient) handleAuthEvent(bytes []byte) {
	slog.Debug("Received auth event")

	if ws.hasAccountHandler() {
		ws.accountEventHandler.handleAuthMessage(bytes)
	}
}

func (ws *wsClient) hasErrorChannel() bool {
	return ws.errchn != nil
}

func (ws *wsClient) hasCandleHandler() bool {
	return ws.candlesEventHandler != nil
}

func (ws *wsClient) hasTickerHandler() bool {
	return ws.tickerEventHandler != nil
}

func (ws *wsClient) hasTicker24hHandler() bool {
	return ws.ticker24hEventHandler != nil
}

func (ws *wsClient) hasTradesHandler() bool {
	return ws.tradesEventHandler != nil
}

func (ws *wsClient) hasBookHandler() bool {
	return ws.bookEventHandler != nil
}

func (ws *wsClient) hasAccountHandler() bool {
	return ws.accountEventHandler != nil
}
