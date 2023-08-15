package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/vladislav-chunikhin/lib-go/pkg/circuitbreaker"
	"github.com/vladislav-chunikhin/lib-go/pkg/limiter"
	"github.com/vladislav-chunikhin/lib-go/pkg/logger"
)

const (
	contentTypeHeader = "Content-Type"
	jsonContent       = "application/json"
	xmlContent        = "text/xml"
)

type Client struct {
	client  *http.Client
	baseURL *url.URL
	logger  logger.Logger
}

func NewClient(
	baseURL string,
	timeout time.Duration,
	logger logger.Logger,
) (*Client, error) {
	URL, err := url.ParseRequestURI(baseURL)
	if err != nil {
		return nil, err
	}

	return &Client{
		client: &http.Client{
			Transport: http.DefaultTransport,
			Timeout:   timeout,
		},
		baseURL: URL,
		logger:  logger,
	}, nil
}

func (c *Client) AddCircuitBreaker(serviceName string, cbCfg *circuitbreaker.Config) {
	c.client.Transport = circuitbreaker.NewProxy(
		serviceName,
		cbCfg,
		c.logger,
		c.client.Transport,
	)
}

func (c *Client) AddLimiter(limit int) {
	c.client.Transport = limiter.NewProxy(
		limit,
		c.client.Transport,
	)
}

func (c *Client) Do(ctx context.Context, req *http.Request, v any) error {
	req = req.WithContext(ctx)

	req.URL = updateRequestURL(c.baseURL, req.URL)

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode < http.StatusOK || res.StatusCode > http.StatusIMUsed {
		return fmt.Errorf("the status code is not in the 2XX range: %d", res.StatusCode)
	}

	if err = decodeJSON(res, v); err != nil {
		return err
	}

	return nil
}

func updateRequestURL(baseURL, targetURL *url.URL) *url.URL {
	result := *targetURL
	result.Host = baseURL.Host
	result.Scheme = baseURL.Scheme
	result.Path = joinURLPaths(baseURL.Path, targetURL.Path)
	result.User = baseURL.User

	return &result
}

func joinURLPaths(p1, p2 string) string {
	p1 = strings.TrimSuffix(p1, "/")
	p2 = strings.TrimPrefix(p2, "/")

	return fmt.Sprintf("%s/%s", p1, p2)
}

func decodeJSON(res *http.Response, v any) (err error) {
	if v == nil || res == nil {
		return nil
	}

	var b bytes.Buffer

	if _, err = io.Copy(&b, res.Body); err != nil {
		return
	}

	defer func() {
		if bodyErr := res.Body.Close(); bodyErr != nil {
			err = bodyErr
		}
	}()

	contentType := strings.ToLower(res.Header.Get(contentTypeHeader))

	switch contentType {
	case jsonContent:
		if jsonErr := json.NewDecoder(&b).Decode(v); jsonErr != nil {
			return jsonErr
		}
		return nil
	case xmlContent:
		if xmlErr := xml.NewDecoder(&b).Decode(v); xmlErr != nil {
			return xmlErr
		}
	default:
		return fmt.Errorf("unexpected content type: %s", contentType)
	}

	return nil
}
