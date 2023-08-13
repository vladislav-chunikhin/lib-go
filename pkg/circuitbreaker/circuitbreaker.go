package circuitbreaker

import (
	"errors"
	"net/http"
	"time"

	"github.com/sony/gobreaker"
)

const (
	metricsCircuitBreakerStateOpen     = 1.0
	metricsCircuitBreakerStateHalfOpen = 0.5
	metricsCircuitBreakerStateClosed   = 0.0
)

var (
	ErrorCircuitBreakerOpened          = errors.New("circuit breaker opened")
	ErrorCircuitBreakerTooManyRequests = errors.New("circuit breaker too many requests")
	ErrorCircuitBreakerInvalidResponse = errors.New("invalid response type from circuit breaker")
)

type CircuitBreakerSettings struct {
	// Minimum number of all requests when calculating failureRatio when you need to change state Closed to Open
	MinRequestsForOpen uint32
	// Proportion of unsuccessful requests when you need to change state Closed to Open
	MinFailureRatioForOpen float64
	// Time after which to reset the counters in the Closed state.
	IntervalResetToClosed time.Duration
	// Time after which to change state Open to Half-open
	TimeoutSwitchToHalfOpen time.Duration
	CallOnResponse          func(*http.Response, error) (*http.Response, error)
}

type TransportWithCircuitBreaker struct {
	cb             *gobreaker.CircuitBreaker
	apiName        string
	nextTransport  http.RoundTripper
	callOnResponse func(*http.Response, error) (*http.Response, error)
}

func (t *TransportWithCircuitBreaker) RoundTrip(req *http.Request) (*http.Response, error) {
	cbResp, err := t.cb.Execute(func() (interface{}, error) {
		// nolint:bodyclose // we don't need to close the body here
		resp, err := t.nextTransport.RoundTrip(req)
		// nolint:bodyclose // it will be wrapped in the caller
		return t.callOnResponse(resp, err)
	})

	if errors.Is(err, gobreaker.ErrOpenState) {
		return nil, ErrorCircuitBreakerOpened
	}

	if errors.Is(err, gobreaker.ErrTooManyRequests) {
		return nil, ErrorCircuitBreakerTooManyRequests
	}

	if err != nil {
		return nil, err
	}

	resp, ok := cbResp.(*http.Response)
	if !ok {
		return nil, ErrorCircuitBreakerInvalidResponse
	}

	return resp, nil
}
