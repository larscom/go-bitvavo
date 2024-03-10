package ws

import (
	"github.com/google/uuid"
	"github.com/orsinium-labs/enum"
	"github.com/smallnest/safemap"
)

type WsEvent enum.Member[string]

var (
	wsEventSubscribed   = WsEvent{"subscribed"}
	wsEventUnsubscribed = WsEvent{"unsubscribed"}
	wsEventCandles      = WsEvent{"candle"}
	wsEventTicker       = WsEvent{"ticker"}
	wsEventTicker24h    = WsEvent{"ticker24h"}
	wsEventTrades       = WsEvent{"trade"}
	wsEventBook         = WsEvent{"book"}
	wsEventAuth         = WsEvent{"authenticate"}
	wsEventOrder        = WsEvent{"order"}
	wsEventFill         = WsEvent{"fill"}
)

type Action enum.Member[string]

var (
	actionSubscribe    = Action{"subscribe"}
	actionUnsubscribe  = Action{"unsubscribe"}
	actionAuthenticate = Action{"authenticate"}
)

type ChannelName enum.Member[string]

var (
	channelNameCandles   = ChannelName{"candles"}
	channelNameTicker    = ChannelName{"ticker"}
	channelNameTicker24h = ChannelName{"ticker24h"}
	channelNameTrades    = ChannelName{"trades"}
	channelNameBook      = ChannelName{"book"}
	channelNameAccount   = ChannelName{"account"}
)

type subscription[T any] struct {
	id     uuid.UUID
	market string

	outchn chan T
	inchn  chan<- T
}

func newSubscription[T any](id uuid.UUID, market string, inchn chan<- T, outchn chan T) *subscription[T] {
	return &subscription[T]{
		id:     id,
		market: market,
		inchn:  inchn,
		outchn: outchn,
	}
}

func relayMessages[T any](in <-chan T, out chan<- T) {
	for msg := range in {
		out <- msg
	}
}

func requireSubscription[T any](subs *safemap.SafeMap[string, T], markets []string) error {
	for _, market := range markets {
		if !subs.Has(market) {
			return errNoSubscriptionActive(market)
		}
	}
	return nil
}

func requireNoSubscription[T any](subs *safemap.SafeMap[string, T], markets []string) error {
	for _, market := range markets {
		if subs.Has(market) {
			return errSubscriptionAlreadyActive(market)
		}
	}
	return nil
}

func deleteSubscriptions[T any](
	subs *safemap.SafeMap[string, *subscription[T]],
	keys []string,
) error {
	counts := make(map[uuid.UUID]int)
	for item := range subs.IterBuffered() {
		counts[item.Val.id]++
	}

	idsWithKeys := make(map[uuid.UUID][]string)
	for _, key := range keys {
		if sub, found := subs.Get(key); found {
			idsWithKeys[sub.id] = append(idsWithKeys[sub.id], key)
			close(sub.inchn)
		}
	}

	for id, key := range idsWithKeys {
		if counts[id] == len(key) {
			if item, found := subs.Get(key[0]); found {
				close(item.outchn)
			}
		}
		for _, key := range key {
			subs.Remove(key)
		}
	}

	return nil
}
