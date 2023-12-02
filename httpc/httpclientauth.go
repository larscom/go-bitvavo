package httpc

import (
	"fmt"
	"time"

	"github.com/larscom/go-bitvavo/v2/jsond"
)

type HttpClientAuth interface {
	// GetBalance returns the balance on the account
	GetBalance() ([]jsond.Balance, error)

	// GetAccount returns trading volume and fees for account
	GetAccount() (jsond.Account, error)
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

func (c *httpClientAuth) GetBalance() ([]jsond.Balance, error) {
	return httpGet[[]jsond.Balance](fmt.Sprintf("%s/balance", httpUrl), c.updateRateLimit, c.updateRateLimitResetAt, c.logDebug, c.config)
}

func (c *httpClientAuth) GetAccount() (jsond.Account, error) {
	return httpGet[jsond.Account](fmt.Sprintf("%s/account", httpUrl), c.updateRateLimit, c.updateRateLimitResetAt, c.logDebug, c.config)
}
