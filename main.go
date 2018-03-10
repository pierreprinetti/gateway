package main // import "github.com/pierreprinetti/gateway"

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pierreprinetti/gateway/httpproxy"
	"github.com/pierreprinetti/gateway/nsqproxy"
	methodmux "github.com/pierreprinetti/go-methodmux"
)

type route struct {
	method, path, target string
}

func newRouterWithEnvConfig() (h http.Handler, err error) {
	mux := methodmux.New()

	for _, route := range configFromEnv("PROXY_") {
		switch {
		case strings.HasPrefix(route.target, "nsq://"):
			h, err = nsqproxy.New(route.target)
		default:
			h, err = httpproxy.New(route.target)
		}

		if err != nil {
			return nil, fmt.Errorf("parsing creating the proxy to %q: %v", route.target, err)
		}

		mux.Handle(route.method, route.path+"/", h)
	}
	return mux, nil
}

func main() {
	router, err := newRouterWithEnvConfig()
	if err != nil {
		log.Fatalf("parsing the configuration: %v", err)
	}

	srv := &http.Server{
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if addr := os.Getenv("HTTP_ADDR"); addr != "" {
		srv.Addr = addr
	}

	log.Printf("Gateway started.")

	log.Fatal(srv.ListenAndServe())
}
