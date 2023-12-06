package httpc

import (
	"bytes"
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
)

var (
	client      = http.DefaultClient
	emptyParams = make(url.Values)
)

func httpGet[T any](
	url string,
	params url.Values,
	updateRateLimit func(ratelimit int64),
	updateRateLimitResetAt func(resetAt time.Time),
	logDebug func(message string, args ...any),
	config *authConfig,
) (T, error) {
	req, _ := http.NewRequest("GET", createRequestUrl(url, params), nil)

	return httpDo[T](req, updateRateLimit, updateRateLimitResetAt, logDebug, config)
}

func httpPost[T any](
	url string,
	body T,
	params url.Values,
	updateRateLimit func(ratelimit int64),
	updateRateLimitResetAt func(resetAt time.Time),
	logDebug func(message string, args ...any),
	config *authConfig,
) (T, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return body, err
	}

	req, _ := http.NewRequest("POST", createRequestUrl(url, params), bytes.NewBuffer(payload))
	return httpDo[T](req, updateRateLimit, updateRateLimitResetAt, logDebug, config)
}

func httpDo[T any](
	request *http.Request,
	updateRateLimit func(ratelimit int64),
	updateRateLimitResetAt func(resetAt time.Time),
	logDebug func(message string, args ...any),
	config *authConfig,
) (T, error) {
	logDebug("executing request", "method", request.Method, "url", request.URL.String())

	var empty T
	if err := applyHeaders(request, config); err != nil {
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

func applyHeaders(request *http.Request, config *authConfig) error {
	if config == nil {
		return nil
	}

	body := make([]byte, 0)
	if request.Body != nil {
		bytes, err := io.ReadAll(request.Body)
		if err != nil {
			return err
		}
		body = append(body, bytes...)
	}
	timestamp := time.Now().UnixMilli()

	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set(headerAccessKey, config.apiKey)
	request.Header.Set(headerAccessSignature, crypto.CreateSignature(request.Method, strings.Replace(request.URL.String(), httpUrl, "", 1), body, timestamp, config.apiSecret))
	request.Header.Set(headerAccessTimestamp, fmt.Sprint(timestamp))
	request.Header.Set(headerAccessWindow, fmt.Sprint(config.windowTimeMs))

	return nil
}

func createRequestUrl(url string, params url.Values) string {
	return util.IfOrElse(len(params) > 0, func() string { return fmt.Sprintf("%s?%s", url, params.Encode()) }, url)
}
