package circuitbreaker

import (
	"net/http"

	"github.com/vladislav-chunikhin/lib-go/pkg/logger"
)

func WrapTransportWithCircuitBreaker(
	serviceName string,
	cfg *Config,
	log logger.Logger,
	nextInterceptor http.RoundTripper,
) http.RoundTripper {
	proxy := NewProxy(
		serviceName,
		cfg,
		log,
		nextInterceptor,
		defaultCallOnResponse,
	)

	return proxy
}

func defaultCallOnResponse(resp *http.Response, err error) (*http.Response, error) {
	return resp, err
}
