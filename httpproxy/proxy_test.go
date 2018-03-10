package httpproxy_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

import (
	. "github.com/pierreprinetti/gateway/httpproxy"
)

func TestNew(t *testing.T) {
	t.Run("returns a reverse proxy handler", func(t *testing.T) {
		upstreamSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(418)
			fmt.Fprint(w, "Hello, client")
		}))
		defer upstreamSrv.Close()

		h, err := New(upstreamSrv.URL)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		req := httptest.NewRequest("GET", "http://example.com/foo", nil)
		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		resp := w.Result()
		body, _ := ioutil.ReadAll(resp.Body)
		if want, have := "Hello, client", string(body); have != want {
			t.Errorf("expected body %q, found %q", want, have)
		}
	})

	t.Run("ErrMissingUrlScheme error if target is missing the scheme", func(t *testing.T) {
		_, have := New("google.com")
		if want := ErrMissingUrlScheme; have != want {
			t.Errorf("expected error `%v`, found `%v`", want, have)
		}
	})

	t.Run("errors if target is not an URL", func(t *testing.T) {
		_, err := New("�http://http://invalid")
		if err == nil {
			t.Errorf("expected error, found `<nil>`")
		}
	})
}
