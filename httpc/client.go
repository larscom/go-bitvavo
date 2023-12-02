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
	"github.com/larscom/go-bitvavo/v2/jsond"
	"github.com/larscom/go-bitvavo/v2/util"
)

var (
	client = http.DefaultClient
)

func httpGet[T any](
	url string,
	params url.Values,
	updateRateLimit func(ratelimit int64),
	updateRateLimitResetAt func(resetAt time.Time),
	logDebug func(message string, args ...any),
	config *authConfig,
) (T, error) {
	reqUrl := util.IfOrElse(len(params) > 0, func() string { return fmt.Sprintf("%s?%s", url, params.Encode()) }, url)
	req, _ := http.NewRequest("GET", reqUrl, nil)

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

	reqUrl := util.IfOrElse(len(params) > 0, func() string { return fmt.Sprintf("%s?%s", url, params.Encode()) }, url)
	req, _ := http.NewRequest("POST", reqUrl, bytes.NewBuffer(payload))
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

	var data T
	if err := applyHeaders(request, config); err != nil {
		return data, err
	}

	response, err := client.Do(request)
	if err != nil {
		return data, err
	}

	for key, value := range response.Header {
		if key == headerRatelimit {
			if len(value) == 0 {
				return data, fmt.Errorf("header: %s didn't contain a value", headerRatelimit)
			}
			updateRateLimit(util.MustInt64(value[0]))
		}
		if key == headerRatelimitResetAt {
			if len(value) == 0 {
				return data, fmt.Errorf("header: %s didn't contain a value", headerRatelimitResetAt)
			}
			updateRateLimitResetAt(time.UnixMilli(util.MustInt64(value[0])))
		}
	}

	if response.StatusCode > http.StatusIMUsed {
		bytes, err := io.ReadAll(response.Body)
		if err != nil {
			return data, err
		}

		var apiError *jsond.BitvavoErr
		if err := json.Unmarshal(bytes, &apiError); err != nil {
			return data, fmt.Errorf("did not get OK response, code=%d, body=%s", response.StatusCode, string(bytes))
		}
		return data, apiError
	}

	defer response.Body.Close()

	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return data, err
	}

	if err := json.Unmarshal(bytes, &data); err != nil {
		return data, err
	}

	return data, nil
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
