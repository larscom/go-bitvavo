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
		symbol               = j["symbol"].(string)
		name                 = j["name"].(string)
		decimals             = j["decimals"].(float64)
		depositFee           = j["depositFee"].(string)
		depositConfirmations = j["depositConfirmations"].(float64)
		depositStatus        = j["depositStatus"].(string)
		withdrawalFee        = j["withdrawalFee"].(string)
		withdrawalMinAmount  = j["withdrawalMinAmount"].(string)
		withdrawalStatus     = j["withdrawalStatus"].(string)
		networksAny          = j["networks"].([]any)
		message              = j["message"].(string)
	)

	networks := make([]string, len(networksAny))
	for i := 0; i < len(networksAny); i++ {
		networks[i] = networksAny[i].(string)
	}

	m.Symbol = symbol
	m.Name = name
	m.Decimals = int64(decimals)
	m.DepositFee = util.MustFloat64(depositFee)
	m.DepositConfirmations = int64(depositConfirmations)
	m.DepositStatus = depositStatus
	m.WithdrawalFee = util.MustFloat64(withdrawalFee)
	m.WithdrawalMinAmount = util.MustFloat64(withdrawalMinAmount)
	m.WithdrawalStatus = withdrawalStatus
	m.Networks = networks
	m.Message = message

	return nil
}
