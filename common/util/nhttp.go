package util

import (
	"encoding/json"
	"errors"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	"strings"
	//"net/http"
	"time"
)

//net/http: TLS handshake timeout
type Client struct {
	RequestPath string
	Query       map[string]string
	Body        map[string]string
	Header      map[string]string
}

func (c *Client) Get() (map[string]interface{}, error) {
	client := resty.New()
	// Retries are configured per client
	client.
		// Set retry count to non zero to enable retries
		SetRetryCount(3).
		// You can override initial retry wait time.
		// Default is 100 milliseconds.
		SetRetryWaitTime(5 * time.Second).
		// MaxWaitTime can be overridden as well.
		// Default is 2 seconds.
		SetRetryMaxWaitTime(20 * time.Second).
		// SetRetryAfter sets callback to calculate wait time between retries.
		// Default (nil) implies exponential backoff with jitter
		SetRetryAfter(func(client *resty.Client, resp *resty.Response) (time.Duration, error) {
			return 0, errors.New("quota exceeded")
		}).SetTimeout(2 * time.Minute).SetRetryWaitTime(5 * time.Second).
		// RetryConditionFunc type is for retry condition function
		// input: non-nil Response OR request execution error
		AddRetryCondition(func(r *resty.Response, err error) bool {
			if err != nil && strings.Contains(err.Error(), "net/http: TLS handshake timeout") {
				logrus.Warn(" Warning: net/http: TLS handshake timeout")
				return true
			}
			return false
		})
	req := client.R().EnableTrace().SetQueryParams(c.Query).SetBody(c.Body).SetHeaders(c.Header).SetHeader("Accept", "application/json")
	resp, err := req.Get(c.RequestPath)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	res := make(map[string]interface{})
	err = json.Unmarshal(resp.Body(), &res)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	return res, nil
}

func (c *Client) GetCommon() (interface{}, error) {
	client := resty.New()
	// Retries are configured per client
	client.
		// Set retry count to non zero to enable retries
		SetRetryCount(3).
		// You can override initial retry wait time.
		// Default is 100 milliseconds.
		SetRetryWaitTime(5 * time.Second).
		// MaxWaitTime can be overridden as well.
		// Default is 2 seconds.
		SetRetryMaxWaitTime(20 * time.Second).
		// SetRetryAfter sets callback to calculate wait time between retries.
		// Default (nil) implies exponential backoff with jitter
		SetRetryAfter(func(client *resty.Client, resp *resty.Response) (time.Duration, error) {
			return 0, errors.New("quota exceeded")
		}).SetTimeout(2 * time.Minute).SetRetryWaitTime(5 * time.Second).
		// RetryConditionFunc type is for retry condition function
		// input: non-nil Response OR request execution error
		AddRetryCondition(func(r *resty.Response, err error) bool {
			if err != nil && strings.Contains(err.Error(), "net/http: TLS handshake timeout") {
				logrus.Warn(" Warning: net/http: TLS handshake timeout")
				return true
			}
			return false
		})
	req := client.R().EnableTrace().SetQueryParams(c.Query).SetBody(c.Body).SetHeaders(c.Header).SetHeader("Accept", "application/json")
	resp, err := req.Get(c.RequestPath)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	return resp.Body(), nil
}
