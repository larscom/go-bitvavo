package jsond

type Page struct {
	// Bid / ask price.
	Price float64 `json:"price"`

	//  Size of 0 means orders are no longer present at that price level, otherwise the returned size is the new total size on that price level.
	Size float64 `json:"size"`
}

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
