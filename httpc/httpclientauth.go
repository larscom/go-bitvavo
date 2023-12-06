package httpc

import (
	"fmt"
	"time"

	"net/url"

	"github.com/larscom/go-bitvavo/v2/types"
)

type HttpClientAuth interface {
	// GetBalance returns the balance on the account.
	// Optionally provide the symbol to filter for in uppercase (e.g: ETH)
	GetBalance(symbol ...string) ([]types.Balance, error)

	// GetAccount returns trading volume and fees for account.
	GetAccount() (types.Account, error)
}

type httpClientAuth struct {
	config                 *authConfig
	updateRateLimit        func(ratelimit int64)
	updateRateLimitResetAt func(resetAt time.Time)
	logDebug               func(message string, args ...any)
}

type authConfig struct {
	apiKey       string
	apiSecret    string
	windowTimeMs uint64
}

func newHttpClientAuth(
	updateRateLimit func(ratelimit int64),
	updateRateLimitResetAt func(resetAt time.Time),
	logDebug func(message string, args ...any),
	config *authConfig,
) *httpClientAuth {
	return &httpClientAuth{
		updateRateLimit:        updateRateLimit,
		updateRateLimitResetAt: updateRateLimitResetAt,
		logDebug:               logDebug,
		config:                 config,
	}
}

func (c *httpClientAuth) GetBalance(symbol ...string) ([]types.Balance, error) {
	params := make(url.Values)
	if len(symbol) > 0 {
		params.Add("symbol", symbol[0])
	}

	return httpGet[[]types.Balance](
		fmt.Sprintf("%s/balance", httpUrl),
		params,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		c.logDebug,
		c.config,
	)
}

func (c *httpClientAuth) GetAccount() (types.Account, error) {
	return httpGet[types.Account](
		fmt.Sprintf("%s/account", httpUrl),
		emptyParams,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		c.logDebug,
		c.config,
	)
}
