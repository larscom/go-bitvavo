package http

import (
	"context"
	"fmt"
	"time"

	"net/url"

	"github.com/larscom/go-bitvavo/v2/types"
)

type HttpClientAuth interface {
	// GetBalance returns the balance on the account.
	// Optionally provide the symbol to filter for in uppercase (e.g: ETH)
	GetBalance(symbol ...string) ([]types.Balance, error)
	GetBalanceWithContext(ctx context.Context, symbol ...string) ([]types.Balance, error)

	// GetAccount returns trading volume and fees for account.
	GetAccount() (types.Account, error)
	GetAccountWithContext(ctx context.Context) (types.Account, error)

	// GetTrades returns historic trades for your account for market (e.g: ETH-EUR)
	//
	// Optionally provide extra params (see: TradeParams)
	GetTrades(market string, params ...OptionalParams) ([]types.TradeHistoric, error)
	GetTradesWithContext(ctx context.Context, market string, params ...OptionalParams) ([]types.TradeHistoric, error)

	// GetOrders returns data for multiple orders at once for market (e.g: ETH-EUR)
	//
	// Optionally provide extra params (see: OrderParams)
	GetOrders(market string, params ...OptionalParams) ([]types.Order, error)
	GetOrdersWithContext(ctx context.Context, market string, params ...OptionalParams) ([]types.Order, error)

	// GetOrdersOpen returns all open orders for market (e.g: ETH-EUR) or all open orders
	// if no market is given.
	GetOrdersOpen(market ...string) ([]types.Order, error)
	GetOrdersOpenWithContext(ctx context.Context, market ...string) ([]types.Order, error)

	// GetOrder returns the order by market and ID
	GetOrder(market string, orderId string) (types.Order, error)
	GetOrderWithContext(ctx context.Context, market string, orderId string) (types.Order, error)

	// CancelOrders cancels multiple orders at once.
	// Either for an entire market (e.g: ETH-EUR) or for the entire account if you
	// omit the market.
	//
	// It returns a slice of orderId's of which are canceled
	CancelOrders(market ...string) ([]string, error)
	CancelOrdersWithContext(ctx context.Context, market ...string) ([]string, error)

	// CancelOrder cancels a single order by ID for the specific market (e.g: ETH-EUR)
	//
	// It returns the canceled orderId if it was canceled
	CancelOrder(market string, orderId string) (string, error)
	CancelOrderWithContext(ctx context.Context, market string, orderId string) (string, error)

	// NewOrder places a new order on the exchange.
	//
	// It returns the new order if it was successfully created
	NewOrder(market string, side string, orderType string, order types.OrderNew) (types.Order, error)
	NewOrderWithContext(ctx context.Context, market string, side string, orderType string, order types.OrderNew) (types.Order, error)

	// UpdateOrder updates an existing order on the exchange.
	//
	// It returns the updated order if it was successfully updated
	UpdateOrder(market string, orderId string, order types.OrderUpdate) (types.Order, error)
	UpdateOrderWithContext(ctx context.Context, market string, orderId string, order types.OrderUpdate) (types.Order, error)

	// GetDepositAsset returns deposit address (with paymentid for some assets)
	// or bank account information to increase your balance for a specific symbol (e.g: ETH)
	GetDepositAsset(symbol string) (types.DepositAsset, error)
	GetDepositAssetWithContext(ctx context.Context, symbol string) (types.DepositAsset, error)

	// GetDepositHistory returns the deposit history of the account.
	//
	// Optionally provide extra params (see: DepositHistoryParams)
	GetDepositHistory(params ...OptionalParams) ([]types.DepositHistory, error)
	GetDepositHistoryWithContext(ctx context.Context, params ...OptionalParams) ([]types.DepositHistory, error)

	// GetWithdrawalHistory returns the withdrawal history of the account.
	//
	// Optionally provide extra params (see: WithdrawalHistoryParams)
	GetWithdrawalHistory(params ...OptionalParams) ([]types.WithdrawalHistory, error)
	GetWithdrawalHistoryWithContext(ctx context.Context, params ...OptionalParams) ([]types.WithdrawalHistory, error)

	// Withdraw requests a withdrawal to an external cryptocurrency address or verified bank account.
	// Please note that 2FA and address confirmation by e-mail are disabled for API withdrawals.
	Withdraw(symbol string, amount float64, address string, withdrawal types.Withdrawal) (types.WithDrawalResponse, error)
	WithdrawWithContext(ctx context.Context, symbol string, amount float64, address string, withdrawal types.Withdrawal) (types.WithDrawalResponse, error)
}

type httpClientAuth struct {
	config                 *authConfig
	updateRateLimit        func(ratelimit int64)
	updateRateLimitResetAt func(resetAt time.Time)
}

type authConfig struct {
	apiKey       string
	apiSecret    string
	windowTimeMs uint64
}

func newHttpClientAuth(
	updateRateLimit func(ratelimit int64),
	updateRateLimitResetAt func(resetAt time.Time),
	config *authConfig,
) *httpClientAuth {
	return &httpClientAuth{
		updateRateLimit:        updateRateLimit,
		updateRateLimitResetAt: updateRateLimitResetAt,
		config:                 config,
	}
}

func (c *httpClientAuth) GetBalance(symbol ...string) ([]types.Balance, error) {
	return c.GetBalanceWithContext(context.Background(), symbol...)
}

func (c *httpClientAuth) GetBalanceWithContext(ctx context.Context, symbol ...string) ([]types.Balance, error) {
	params := make(url.Values)
	if len(symbol) > 0 {
		params.Add("symbol", symbol[0])
	}

	return httpGet[[]types.Balance](
		ctx,
		fmt.Sprintf("%s/balance", bitvavoURL),
		params,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		c.config,
	)
}

func (c *httpClientAuth) GetAccount() (types.Account, error) {
	return c.GetAccountWithContext(context.Background())
}

func (c *httpClientAuth) GetAccountWithContext(ctx context.Context) (types.Account, error) {
	return httpGet[types.Account](
		ctx,
		fmt.Sprintf("%s/account", bitvavoURL),
		emptyParams,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		c.config,
	)
}

func (c *httpClientAuth) GetOrders(market string, opt ...OptionalParams) ([]types.Order, error) {
	return c.GetOrdersWithContext(context.Background(), market, opt...)
}

func (c *httpClientAuth) GetOrdersWithContext(ctx context.Context, market string, opt ...OptionalParams) ([]types.Order, error) {
	params := make(url.Values)
	if len(opt) > 0 {
		params = opt[0].Params()
	}
	params.Add("market", market)

	return httpGet[[]types.Order](
		ctx,
		fmt.Sprintf("%s/orders", bitvavoURL),
		params,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		c.config,
	)
}

func (c *httpClientAuth) GetOrdersOpen(market ...string) ([]types.Order, error) {
	return c.GetOrdersOpenWithContext(context.Background(), market...)
}

func (c *httpClientAuth) GetOrdersOpenWithContext(ctx context.Context, market ...string) ([]types.Order, error) {
	params := make(url.Values)
	if len(market) > 0 {
		params.Add("market", market[0])
	}

	return httpGet[[]types.Order](
		ctx,
		fmt.Sprintf("%s/ordersOpen", bitvavoURL),
		params,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		c.config,
	)
}

func (c *httpClientAuth) GetOrder(market string, orderId string) (types.Order, error) {
	return c.GetOrderWithContext(context.Background(), market, orderId)
}

func (c *httpClientAuth) GetOrderWithContext(ctx context.Context, market string, orderId string) (types.Order, error) {
	params := make(url.Values)
	params.Add("market", market)
	params.Add("orderId", orderId)

	return httpGet[types.Order](
		ctx,
		fmt.Sprintf("%s/order", bitvavoURL),
		params,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		c.config,
	)
}

func (c *httpClientAuth) CancelOrders(market ...string) ([]string, error) {
	return c.CancelOrdersWithContext(context.Background(), market...)
}

func (c *httpClientAuth) CancelOrdersWithContext(ctx context.Context, market ...string) ([]string, error) {
	params := make(url.Values)
	if len(market) > 0 {
		params.Add("market", market[0])
	}

	resp, err := httpDelete[[]map[string]string](
		ctx,
		fmt.Sprintf("%s/orders", bitvavoURL),
		params,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		c.config,
	)
	if err != nil {
		return nil, err
	}

	orderIds := make([]string, len(resp))
	for i := 0; i < len(orderIds); i++ {
		orderIds[i] = resp[i]["orderId"]
	}

	return orderIds, nil
}

func (c *httpClientAuth) CancelOrder(market string, orderId string) (string, error) {
	return c.CancelOrderWithContext(context.Background(), market, orderId)
}

func (c *httpClientAuth) CancelOrderWithContext(ctx context.Context, market string, orderId string) (string, error) {
	params := make(url.Values)
	params.Add("market", market)
	params.Add("orderId", orderId)

	resp, err := httpDelete[map[string]string](
		ctx,
		fmt.Sprintf("%s/order", bitvavoURL),
		params,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		c.config,
	)
	if err != nil {
		return "", err
	}

	return resp["orderId"], nil
}

func (c *httpClientAuth) NewOrder(market string, side string, orderType string, order types.OrderNew) (types.Order, error) {
	return c.NewOrderWithContext(context.Background(), market, side, orderType, order)
}

func (c *httpClientAuth) NewOrderWithContext(ctx context.Context, market string, side string, orderType string, order types.OrderNew) (types.Order, error) {
	order.Market = market
	order.Side = side
	order.OrderType = orderType
	return httpPost[types.Order](
		ctx,
		fmt.Sprintf("%s/order", bitvavoURL),
		order,
		emptyParams,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		c.config,
	)
}

func (c *httpClientAuth) UpdateOrder(market string, orderId string, order types.OrderUpdate) (types.Order, error) {
	return c.UpdateOrderWithContext(context.Background(), market, orderId, order)
}

func (c *httpClientAuth) UpdateOrderWithContext(ctx context.Context, market string, orderId string, order types.OrderUpdate) (types.Order, error) {
	order.Market = market
	order.OrderId = orderId

	return httpPut[types.Order](
		ctx,
		fmt.Sprintf("%s/order", bitvavoURL),
		order,
		emptyParams,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		c.config,
	)
}

func (c *httpClientAuth) GetTrades(market string, opt ...OptionalParams) ([]types.TradeHistoric, error) {
	return c.GetTradesWithContext(context.Background(), market, opt...)
}

func (c *httpClientAuth) GetTradesWithContext(ctx context.Context, market string, opt ...OptionalParams) ([]types.TradeHistoric, error) {
	params := make(url.Values)
	if len(opt) > 0 {
		params = opt[0].Params()
	}
	params.Add("market", market)

	return httpGet[[]types.TradeHistoric](
		ctx,
		fmt.Sprintf("%s/trades", bitvavoURL),
		params,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		c.config,
	)
}

func (c *httpClientAuth) GetDepositAsset(symbol string) (types.DepositAsset, error) {
	return c.GetDepositAssetWithContext(context.Background(), symbol)
}

func (c *httpClientAuth) GetDepositAssetWithContext(ctx context.Context, symbol string) (types.DepositAsset, error) {
	params := make(url.Values)
	params.Add("symbol", symbol)

	return httpGet[types.DepositAsset](
		ctx,
		fmt.Sprintf("%s/deposit", bitvavoURL),
		params,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		c.config,
	)
}

func (c *httpClientAuth) GetDepositHistory(opt ...OptionalParams) ([]types.DepositHistory, error) {
	return c.GetDepositHistoryWithContext(context.Background(), opt...)
}

func (c *httpClientAuth) GetDepositHistoryWithContext(ctx context.Context, opt ...OptionalParams) ([]types.DepositHistory, error) {
	params := make(url.Values)
	if len(opt) > 0 {
		params = opt[0].Params()
	}
	return httpGet[[]types.DepositHistory](
		ctx,
		fmt.Sprintf("%s/depositHistory", bitvavoURL),
		params,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		c.config,
	)
}

func (c *httpClientAuth) GetWithdrawalHistory(opt ...OptionalParams) ([]types.WithdrawalHistory, error) {
	return c.GetWithdrawalHistoryWithContext(context.Background(), opt...)
}

func (c *httpClientAuth) GetWithdrawalHistoryWithContext(ctx context.Context, opt ...OptionalParams) ([]types.WithdrawalHistory, error) {
	params := make(url.Values)
	if len(opt) > 0 {
		params = opt[0].Params()
	}
	return httpGet[[]types.WithdrawalHistory](
		ctx,
		fmt.Sprintf("%s/withdrawalHistory", bitvavoURL),
		params,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		c.config,
	)
}

func (c *httpClientAuth) Withdraw(symbol string, amount float64, address string, withdrawal types.Withdrawal) (types.WithDrawalResponse, error) {
	return c.WithdrawWithContext(context.Background(), symbol, amount, address, withdrawal)
}

func (c *httpClientAuth) WithdrawWithContext(ctx context.Context, symbol string, amount float64, address string, withdrawal types.Withdrawal) (types.WithDrawalResponse, error) {
	withdrawal.Symbol = symbol
	withdrawal.Amount = amount
	withdrawal.Address = address

	return httpPost[types.WithDrawalResponse](
		ctx,
		fmt.Sprintf("%s/withdrawal", bitvavoURL),
		withdrawal,
		emptyParams,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		c.config,
	)
}
