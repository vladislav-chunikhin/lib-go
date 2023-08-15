package limiter

import (
	"net/http"

	"golang.org/x/time/rate"
)

type Config struct {
	Limit int `yaml:"limit"`
}

type Proxy struct {
	limiter       *rate.Limiter
	nextTransport http.RoundTripper
}

func NewProxy(limit int, nextInterceptor http.RoundTripper) *Proxy {
	return &Proxy{
		limiter:       rate.NewLimiter(rate.Limit(limit), limit),
		nextTransport: nextInterceptor,
	}
}

func (p *Proxy) RoundTrip(req *http.Request) (*http.Response, error) {
	if !p.limiter.Allow() {
		p.limiter.Wait(req.Context())
	}

	return p.nextTransport.RoundTrip(req)
}
