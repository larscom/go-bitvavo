package types

import (
	"github.com/goccy/go-json"
	"github.com/larscom/go-bitvavo/v2/util"
)

type Asset struct {
	// Short version of the asset name used in market names.
	Symbol string `json:"symbol"`

	// The full name of the asset.
	Name string `json:"name"`

	// The precision used for specifying amounts.
	Decimals int64 `json:"decimals"`

	// Fixed fee for depositing this asset.
	DepositFee float64 `json:"depositFee"`

	// The minimum amount of network confirmations required before this asset is credited to your account.
	DepositConfirmations int64 `json:"depositConfirmations"`

	// Enum: "OK" | "MAINTENANCE" | "DELISTED"
	DepositStatus string `json:"depositStatus"`

	// Fixed fee for withdrawing this asset.
	WithdrawalFee float64 `json:"withdrawalFee"`

	// The minimum amount for which a withdrawal can be made.
	WithdrawalMinAmount float64 `json:"withdrawalMinAmount"`

	// Enum: "OK" | "MAINTENANCE" | "DELISTED"
	WithdrawalStatus string `json:"withdrawalStatus"`

	// Supported networks.
	Networks []string `json:"networks"`

	// Shows the reason if withdrawalStatus or depositStatus is not OK.
	Message string `json:"message"`
}

func (m *Asset) UnmarshalJSON(bytes []byte) error {
	var j map[string]any

	err := json.Unmarshal(bytes, &j)
	if err != nil {
		return err
	}

	var (
		symbol               = getOrEmpty[string]("symbol", j)
		name                 = getOrEmpty[string]("name", j)
		decimals             = getOrEmpty[float64]("decimals", j)
		depositFee           = getOrEmpty[string]("depositFee", j)
		depositConfirmations = getOrEmpty[float64]("depositConfirmations", j)
		depositStatus        = getOrEmpty[string]("depositStatus", j)
		withdrawalFee        = getOrEmpty[string]("withdrawalFee", j)
		withdrawalMinAmount  = getOrEmpty[string]("withdrawalMinAmount", j)
		withdrawalStatus     = getOrEmpty[string]("withdrawalStatus", j)
		networksAny          = getOrEmpty[[]any]("networks", j)
		message              = getOrEmpty[string]("message", j)
	)

	networks := make([]string, len(networksAny))
	for i := 0; i < len(networksAny); i++ {
		networks[i] = networksAny[i].(string)
	}

	m.Symbol = symbol
	m.Name = name
	m.Decimals = int64(decimals)
	m.DepositFee = util.IfOrElse(len(depositFee) > 0, func() float64 { return util.MustFloat64(depositFee) }, 0)
	m.DepositConfirmations = int64(depositConfirmations)
	m.DepositStatus = depositStatus
	m.WithdrawalFee = util.IfOrElse(len(withdrawalFee) > 0, func() float64 { return util.MustFloat64(withdrawalFee) }, 0)
	m.WithdrawalMinAmount = util.IfOrElse(len(withdrawalMinAmount) > 0, func() float64 { return util.MustFloat64(withdrawalMinAmount) }, 0)
	m.WithdrawalStatus = withdrawalStatus
	m.Networks = networks
	m.Message = message

	return nil
}
