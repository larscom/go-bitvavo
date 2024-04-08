package ws

import (
	"github.com/google/uuid"
	csmap "github.com/mhmtszr/concurrent-swiss-map"
	"github.com/orsinium-labs/enum"
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
	wsEvents            = enum.New(
		wsEventSubscribed,
		wsEventUnsubscribed,
		wsEventCandles,
		wsEventTicker,
		wsEventTicker24h,
		wsEventTrades,
		wsEventBook,
		wsEventAuth,
		wsEventOrder,
		wsEventFill,
	)
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

func getSubscriptionKeys[K comparable, V any](data *csmap.CsMap[K, V]) []K {
	keys := make([]K, 0)
	data.Range(func(key K, value V) (stop bool) {
		keys = append(keys, key)
		return false
	})
	return keys
}

func relayMessages[T any](in <-chan T, out chan<- T) {
	for msg := range in {
		out <- msg
	}
}

func requireSubscription[T any](subs *csmap.CsMap[string, T], markets []string) error {
	for _, market := range markets {
		if !subs.Has(market) {
			return errNoSubscriptionActive(market)
		}
	}
	return nil
}

func requireNoSubscription[T any](subs *csmap.CsMap[string, T], markets []string) error {
	for _, market := range markets {
		if subs.Has(market) {
			return errSubscriptionAlreadyActive(market)
		}
	}
	return nil
}

func deleteSubscriptions[T any](
	subs *csmap.CsMap[string, *subscription[T]],
	keys []string,
) error {
	counts := make(map[uuid.UUID]int)
	subs.Range(func(key string, value *subscription[T]) (stop bool) {
		counts[value.id]++
		return false
	})

	idsWithKeys := make(map[uuid.UUID][]string)
	for _, key := range keys {
		if sub, found := subs.Load(key); found {
			idsWithKeys[sub.id] = append(idsWithKeys[sub.id], key)
			close(sub.inchn)
		}
	}

	for id, keys := range idsWithKeys {
		if counts[id] == len(keys) {
			if item, found := subs.Load(keys[0]); found {
				close(item.outchn)
			}
		}
		for _, key := range keys {
			subs.Delete(key)
		}
	}

	return nil
}
