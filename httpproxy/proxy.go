// Package httpproxy forwards HTTP calls to an HTTP upstream server.
package httpproxy

import (
	"errors"
	"net/http/httputil"
	"net/url"
)

var ErrMissingUrlScheme = errors.New("target URL is missing the scheme")

// New returns a new httputil.ReverseProxy that routes URLs to the
// scheme, host, and base path provided in target. If the target's
// path is "/base" and the incoming request was for "/dir", the
// target request will be for /base/dir. The returned proxy does not
// rewrite the Host header.
//
// An ErrMissingUrlScheme error will be returned if the provided
// target does not specify a scheme.
func New(target string) (*httputil.ReverseProxy, error) {
	targetURL, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	if targetURL.Scheme == "" {
		return nil, ErrMissingUrlScheme
	}

	return httputil.NewSingleHostReverseProxy(targetURL), nil
}
