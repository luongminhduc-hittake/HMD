package util

import (
	"net/http"
	"time"
)

// SharedTransport provides a single connection pool with keep-alive,
// system proxy support, and reasonable connection timeouts.
var SharedTransport = func() *http.Transport {
	var tr *http.Transport
	if dt, ok := http.DefaultTransport.(*http.Transport); ok {
		tr = dt.Clone()
	} else {
		tr = &http.Transport{}
	}
	tr.MaxIdleConns = 100
	tr.MaxIdleConnsPerHost = 10
	tr.IdleConnTimeout = 90 * time.Second
	tr.ResponseHeaderTimeout = 30 * time.Second
	return tr
}()

// SharedClient is a reusable HTTP client with connection pooling and standard timeout.
var SharedClient = &http.Client{
	Transport: SharedTransport,
	Timeout:   45 * time.Second,
}

// SharedNoRedirectClient is a reusable HTTP client that halts on the first redirect.
var SharedNoRedirectClient = &http.Client{
	Transport: SharedTransport,
	Timeout:   15 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}
