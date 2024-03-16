package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/larscom/go-bitvavo/v2/crypto"
	"github.com/larscom/go-bitvavo/v2/types"
	"github.com/larscom/go-bitvavo/v2/util"
	"github.com/rs/zerolog/log"
)

type OptionalParams interface {
	Params() url.Values
}

var (
	client      = http.DefaultClient
	emptyParams = make(url.Values)
	emptyBody   = make([]byte, 0)
)

func httpDelete[T any](
	ctx context.Context,
	url string,
	params url.Values,
	updateRateLimit func(ratelimit int64),
	updateRateLimitResetAt func(resetAt time.Time),
	config *authConfig,
) (T, error) {
	req, _ := http.NewRequestWithContext(ctx, "DELETE", createRequestUrl(url, params), nil)
	return httpDo[T](req, emptyBody, updateRateLimit, updateRateLimitResetAt, config)
}

func httpGet[T any](
	ctx context.Context,
	url string,
	params url.Values,
	updateRateLimit func(ratelimit int64),
	updateRateLimitResetAt func(resetAt time.Time),
	config *authConfig,
) (T, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", createRequestUrl(url, params), nil)
	return httpDo[T](req, emptyBody, updateRateLimit, updateRateLimitResetAt, config)
}

func httpPost[T any](
	ctx context.Context,
	url string,
	body any,
	params url.Values,
	updateRateLimit func(ratelimit int64),
	updateRateLimitResetAt func(resetAt time.Time),
	config *authConfig,
) (T, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		var empty T
		return empty, err
	}
	log.Debug().Str("body", string(payload)).Msg("created request body")

	req, _ := http.NewRequestWithContext(ctx, "POST", createRequestUrl(url, params), bytes.NewBuffer(payload))
	return httpDo[T](req, payload, updateRateLimit, updateRateLimitResetAt, config)
}

func httpPut[T any](
	ctx context.Context,
	url string,
	body any,
	params url.Values,
	updateRateLimit func(ratelimit int64),
	updateRateLimitResetAt func(resetAt time.Time),
	config *authConfig,
) (T, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		var empty T
		return empty, err
	}
	log.Debug().Str("body", string(payload)).Msg("created request body")

	req, _ := http.NewRequestWithContext(ctx, "PUT", createRequestUrl(url, params), bytes.NewBuffer(payload))
	return httpDo[T](req, payload, updateRateLimit, updateRateLimitResetAt, config)
}

func httpDo[T any](
	request *http.Request,
	body []byte,
	updateRateLimit func(ratelimit int64),
	updateRateLimitResetAt func(resetAt time.Time),
	config *authConfig,
) (T, error) {
	log.Debug().Str("method", request.Method).Str("url", request.URL.String()).Msg("executing request")

	var empty T
	if err := applyHeaders(request, body, config); err != nil {
		return empty, err
	}

	response, err := client.Do(request)
	if err != nil {
		return empty, err
	}
	defer response.Body.Close()

	if err := updateRateLimits(response, updateRateLimit, updateRateLimitResetAt); err != nil {
		return empty, err
	}

	if response.StatusCode > http.StatusIMUsed {
		return empty, unwrapErr(response)
	}

	return unwrapBody[T](response)
}

func unwrapBody[T any](response *http.Response) (T, error) {
	var data T
	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return data, err
	}
	log.Debug().Str("body", string(bytes)).Msg("received response")

	if err := json.Unmarshal(bytes, &data); err != nil {
		return data, err
	}

	return data, nil
}

func unwrapErr(response *http.Response) error {
	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	var bitvavoErr *types.BitvavoErr
	if err := json.Unmarshal(bytes, &bitvavoErr); err != nil {
		return fmt.Errorf("did not get OK response, code=%d, body=%s", response.StatusCode, string(bytes))
	}
	return bitvavoErr
}

func updateRateLimits(
	response *http.Response,
	updateRateLimit func(ratelimit int64),
	updateRateLimitResetAt func(resetAt time.Time),
) error {
	for key, value := range response.Header {
		if key == headerRatelimit {
			if len(value) == 0 {
				return fmt.Errorf("header: %s didn't contain a value", headerRatelimit)
			}
			updateRateLimit(util.MustInt64(value[0]))
		}
		if key == headerRatelimitResetAt {
			if len(value) == 0 {
				return fmt.Errorf("header: %s didn't contain a value", headerRatelimitResetAt)
			}
			updateRateLimitResetAt(time.UnixMilli(util.MustInt64(value[0])))
		}
	}
	return nil
}

func applyHeaders(request *http.Request, body []byte, config *authConfig) error {
	if config == nil {
		return nil
	}

	timestamp := time.Now().UnixMilli()

	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set(headerAccessKey, config.apiKey)
	request.Header.Set(headerAccessSignature, crypto.CreateSignature(request.Method, strings.Replace(request.URL.String(), bitvavoURL, "", 1), body, timestamp, config.apiSecret))
	request.Header.Set(headerAccessTimestamp, fmt.Sprint(timestamp))
	request.Header.Set(headerAccessWindow, fmt.Sprint(config.windowTimeMs))

	return nil
}

func createRequestUrl(url string, params url.Values) string {
	return util.IfOrElse(len(params) > 0, func() string { return fmt.Sprintf("%s?%s", url, params.Encode()) }, url)
}
