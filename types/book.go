package types

import (
	"github.com/goccy/go-json"
	"github.com/larscom/go-bitvavo/v2/util"
)

type Book struct {
	// Integer which is increased by one for every update to the book. Useful for synchronizing. Resets to zero after restarting the matching engine.
	Nonce int64 `json:"nonce"`

	// Slice with all bids in the format [price, size], where an size of 0 means orders are no longer present at that price level,
	// otherwise the returned size is the new total size on that price level.
	Bids []Page `json:"bids"`

	// Slice with all asks in the format [price, size], where an size of 0 means orders are no longer present at that price level,
	// otherwise the returned size is the new total size on that price level.
	Asks []Page `json:"asks"`
}

type Page struct {
	// Bid / ask price.
	Price float64 `json:"price"`

	//  Size of 0 means orders are no longer present at that price level, otherwise the returned size is the new total size on that price level.
	Size float64 `json:"size"`
}

func (b *Book) UnmarshalJSON(bytes []byte) error {
	var j map[string]any

	err := json.Unmarshal(bytes, &j)
	if err != nil {
		return err
	}

	nonce := j["nonce"].(float64)
	bidEvents := j["bids"].([]any)
	askEvents := j["asks"].([]any)

	bids := make([]Page, len(bidEvents))
	for i := 0; i < len(bidEvents); i++ {
		price := bidEvents[i].([]any)[0].(string)
		size := bidEvents[i].([]any)[1].(string)

		bids[i] = Page{
			Price: util.IfOrElse(len(price) > 0, func() float64 { return util.MustFloat64(price) }, 0),
			Size:  util.IfOrElse(len(size) > 0, func() float64 { return util.MustFloat64(size) }, 0),
		}
	}

	asks := make([]Page, len(askEvents))
	for i := 0; i < len(askEvents); i++ {
		price := askEvents[i].([]any)[0].(string)
		size := askEvents[i].([]any)[1].(string)

		asks[i] = Page{
			Price: util.IfOrElse(len(price) > 0, func() float64 { return util.MustFloat64(price) }, 0),
			Size:  util.IfOrElse(len(size) > 0, func() float64 { return util.MustFloat64(size) }, 0),
		}
	}

	b.Nonce = int64(nonce)
	b.Bids = bids
	b.Asks = asks

	return nil
}
