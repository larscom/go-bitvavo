package wsc

import (
	"fmt"

	"github.com/larscom/go-bitvavo/v2/constant"
	"github.com/larscom/go-bitvavo/v2/jsond"
	"github.com/larscom/go-bitvavo/v2/log"

	"github.com/goccy/go-json"
	"github.com/larscom/go-bitvavo/v2/util"
	"github.com/smallnest/safemap"
)

type BookEvent struct {
	// Describes the returned event over the socket.
	Event string `json:"event"`

	// The market which was requested in the subscription.
	Market string `json:"market"`

	// The book containing the bids and asks.
	Book jsond.Book `json:"book"`
}

func (b *BookEvent) UnmarshalJSON(bytes []byte) error {
	var bookEvent map[string]any
	err := json.Unmarshal(bytes, &bookEvent)
	if err != nil {
		return err
	}

	var (
		event  = bookEvent["event"].(string)
		market = bookEvent["market"].(string)
		nonce  = bookEvent["nonce"].(float64)
	)

	bidEvents := bookEvent["bids"].([]any)
	bids := make([]jsond.Page, len(bidEvents))
	for i := 0; i < len(bidEvents); i++ {
		price := bidEvents[i].([]any)[0].(string)
		size := bidEvents[i].([]any)[1].(string)

		bids[i] = jsond.Page{
			Price: util.IfOrElse(len(price) > 0, func() float64 { return util.MustFloat64(price) }, constant.ZEROF64),
			Size:  util.IfOrElse(len(size) > 0, func() float64 { return util.MustFloat64(size) }, constant.ZEROF64),
		}
	}

	askEvents := bookEvent["asks"].([]any)
	asks := make([]jsond.Page, len(askEvents))
	for i := 0; i < len(askEvents); i++ {
		price := askEvents[i].([]any)[0].(string)
		size := askEvents[i].([]any)[1].(string)

		asks[i] = jsond.Page{
			Price: util.IfOrElse(len(price) > 0, func() float64 { return util.MustFloat64(price) }, constant.ZEROF64),
			Size:  util.IfOrElse(len(size) > 0, func() float64 { return util.MustFloat64(size) }, constant.ZEROF64),
		}
	}

	b.Event = event
	b.Market = market
	b.Book = jsond.Book{
		Nonce: int64(nonce),
		Bids:  bids,
		Asks:  asks,
	}

	return nil
}

type bookEventHandler struct {
	writechn chan<- WebSocketMessage
	subs     *safemap.SafeMap[string, chan<- BookEvent]
}

func newBookEventHandler(writechn chan<- WebSocketMessage) *bookEventHandler {
	return &bookEventHandler{
		writechn: writechn,
		subs:     safemap.New[string, chan<- BookEvent](),
	}
}

func (t *bookEventHandler) Subscribe(market string, buffSize uint64) (<-chan BookEvent, error) {
	if t.subs.Has(market) {
		return nil, fmt.Errorf("subscription already active for market: %s", market)
	}

	t.writechn <- newWebSocketMessage(actionSubscribe, channelNameBook, market)

	chn := make(chan BookEvent, buffSize)
	t.subs.Set(market, chn)

	return chn, nil
}

func (t *bookEventHandler) Unsubscribe(market string) error {
	sub, exist := t.subs.Get(market)

	if exist {
		t.writechn <- newWebSocketMessage(actionUnsubscribe, channelNameBook, market)
		close(sub)
		t.subs.Remove(market)
		return nil
	}

	return fmt.Errorf("no subscription active for market: %s", market)
}

func (t *bookEventHandler) UnsubscribeAll() error {
	for sub := range t.subs.IterBuffered() {
		market := sub.Key
		if err := t.Unsubscribe(market); err != nil {
			return err
		}
	}
	return nil
}

func (t *bookEventHandler) handleMessage(bytes []byte) {
	var bookEvent *BookEvent
	if err := json.Unmarshal(bytes, &bookEvent); err != nil {
		log.Logger().Error("Couldn't unmarshal message into BookEvent", "message", string(bytes))
	} else {
		market := bookEvent.Market
		chn, exist := t.subs.Get(market)
		if exist {
			chn <- *bookEvent
		} else {
			log.Logger().Error("There is no active subscription", "handler", "trades", "market", market)
		}
	}
}

func (t *bookEventHandler) reconnect() {
	for sub := range t.subs.IterBuffered() {
		market := sub.Key
		t.writechn <- newWebSocketMessage(actionSubscribe, channelNameBook, market)
	}
}
