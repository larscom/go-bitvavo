package types

import (
	"fmt"
	"net/url"
	"time"

	"github.com/goccy/go-json"
	"github.com/larscom/go-bitvavo/v2/util"
)

type WithdrawalHistoryParams struct {
	// When no symbol is specified, all withdrawal will be returned.
	Symbol string `json:"symbol"`

	// Return the limit most recent assets only.
	// Default: 500
	Limit uint64 `json:"limit"`

	// Return orders after start time.
	Start time.Time `json:"start"`

	// Return orders before end time.
	End time.Time `json:"end"`
}

func (w *WithdrawalHistoryParams) Params() url.Values {
	params := make(url.Values)

	if w.Symbol != "" {
		params.Add("symbol", fmt.Sprint(w.Symbol))
	}
	if w.Limit > 0 {
		params.Add("limit", fmt.Sprint(w.Limit))
	}
	if !w.Start.IsZero() {
		params.Add("start", fmt.Sprint(w.Start.UnixMilli()))
	}
	if !w.End.IsZero() {
		params.Add("end", fmt.Sprint(w.End.UnixMilli()))
	}

	return params
}

type WithdrawalHistory struct {
	// The time your withdrawal of symbol was received by Bitvavo.
	Timestamp int64 `json:"timestamp"`

	// The short name of the asset. For example, BTC for Bitcoin.
	Symbol string `json:"symbol"`

	// Amount that has been withdrawn.
	Amount float64 `json:"amount"`

	// Address that has been used for this withdrawal.
	Address string `json:"address"`

	// Payment ID used for this withdrawal. This is mostly called a note, memo or tag. Will not be returned if it was not used.
	PaymentId string `json:"paymentId"`

	// The transaction ID, which can be found on the blockchain, for this specific withdrawal.
	TxId string `json:"txId"`

	// The fee which has been paid to withdraw this currency.
	Fee float64 `json:"fee"`

	// The status of the withdrawal.
	// Enum: "awaiting_processing" | "awaiting_email_confirmation" | "awaiting_bitvavo_inspection" |
	// "approved" | "sending" | "in_mempool" | "processed" | "completed" | "canceled"
	Status string `json:"status"`
}

func (w *WithdrawalHistory) UnmarshalJSON(bytes []byte) error {
	var j map[string]any

	if err := json.Unmarshal(bytes, &j); err != nil {
		return err
	}

	var (
		timestamp = getOrEmpty[float64]("timestamp", j)
		symbol    = getOrEmpty[string]("symbol", j)
		amount    = getOrEmpty[string]("amount", j)
		address   = getOrEmpty[string]("address", j)
		paymentId = getOrEmpty[string]("paymentId", j)
		txId      = getOrEmpty[string]("txId", j)
		fee       = getOrEmpty[string]("fee", j)
		status    = getOrEmpty[string]("status", j)
	)

	w.Timestamp = int64(timestamp)
	w.Symbol = symbol
	w.Amount = util.IfOrElse(len(amount) > 0, func() float64 { return util.MustFloat64(amount) }, 0)
	w.Address = address
	w.PaymentId = paymentId
	w.TxId = txId
	w.Fee = util.IfOrElse(len(fee) > 0, func() float64 { return util.MustFloat64(fee) }, 0)
	w.Status = status

	return nil
}

type Withdrawal struct {
	// The short name of the asset. For example, BTC for Bitcoin.
	Symbol string `json:"symbol"`

	// Total amount that has been deducted from your balance.
	Amount float64 `json:"amount"`

	// Wallet address or IBAN.
	// For digital assets: please double check this address. Funds sent can not be recovered.
	Address string `json:"address"`

	// For digital assets only. Payment IDs are used to identify transactions to merchants and exchanges with a single address. This is mostly called a note, memo or tag. Should be set when withdrawing straight to another exchange or merchants that require payment id's.
	PaymentId string `json:"paymentId,omitempty"`

	// For digital assets only.
	// Should be set to true if the withdrawal must be sent to another Bitvavo user internally.
	// No transaction will be broadcast to the blockchain and no fees will be applied.
	// This operation fails if the wallet does not belong to a Bitvavo user.
	Internal bool `json:"internal,omitempty"`

	// If set to true, the fee will be added on top of the requested amount,
	// otherwise the fee is part of the requested amount and subtracted from the withdrawal.
	AddWithdrawalFee bool `json:"addWithdrawalFee,omitempty"`
}

type WithDrawalResponse struct {
	// Returns true for successful withdrawal requests.
	Success bool `json:"success"`

	// The short name of the asset. For example, BTC for Bitcoin.
	Symbol string `json:"symbol"`

	// Total amount that has been deducted from your balance.
	Amount float64 `json:"amount"`
}

func (r *WithDrawalResponse) UnmarshalJSON(bytes []byte) error {
	var j map[string]any

	if err := json.Unmarshal(bytes, &j); err != nil {
		return err
	}

	var (
		success = getOrEmpty[bool]("success", j)
		symbol  = getOrEmpty[string]("symbol", j)
		amount  = getOrEmpty[string]("amount", j)
	)

	r.Success = success
	r.Symbol = symbol
	r.Amount = util.IfOrElse(len(amount) > 0, func() float64 { return util.MustFloat64(amount) }, 0)

	return nil
}
