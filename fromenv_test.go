package main

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestRouteFromEnvVar(t *testing.T) {
	testCases := [...]struct {
		name        string
		prefix      string
		env         string
		expectedOut route
		expectedOK  bool
	}{
		{
			"parses without prefix",
			"",
			"GET_google=https://www.google.com",
			route{method: "GET", path: "/google", target: "https://www.google.com"},
			true,
		},
		{
			"parses with prefix",
			"PROXY_",
			"PROXY_GET_google=https://www.google.com",
			route{method: "GET", path: "/google", target: "https://www.google.com"},
			true,
		},
		{
			"returns false if missing prefix",
			"PROXY_",
			"GET_google=https://www.google.com",
			route{},
			false,
		},
		{
			"returns false if not parsable",
			"",
			"GET-google=https://www.google.com",
			route{},
			false,
		},
		{
			"returns false if missing the target",
			"",
			"GET_google=",
			route{},
			false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r, ok := routeFromEnvVar(tc.prefix, tc.env)

			if have, want := r.method, tc.expectedOut.method; have != want {
				t.Errorf("expected method to be %q, found %q", want, have)
			}
			if have, want := r.path, tc.expectedOut.path; have != want {
				t.Errorf("expected path to be %q, found %q", want, have)
			}
			if have, want := r.target, tc.expectedOut.target; have != want {
				t.Errorf("expected target to be %q, found %q", want, have)
			}
			if ok != tc.expectedOK {
				t.Errorf("expected OK value to be %v, found %v", tc.expectedOK, ok)
			}
		})
	}
}

func TestConfigFromEnv(t *testing.T) {
	setenv := func(envVars ...string) func(context.Context) {
		return func(ctx context.Context) {
			for _, env := range envVars {
				envKV := strings.SplitN(env, "=", 2)
				os.Setenv(envKV[0], envKV[1])
			}

			// teardown
			go func() {
				<-ctx.Done()
				for _, env := range envVars {
					envKV := strings.SplitN(env, "=", 2)
					os.Unsetenv(envKV[0])
				}
			}()
		}
	}

	testCases := [...]struct {
		name           string
		prefix         string
		setup          func(context.Context)
		expectedRoutes []route
	}{
		{
			"parses the environment with prefix",
			"prefix",
			setenv(
				"prefixGET_users=https://users.com/",
				"prefixPOST_x=http://user_service:8081",
			),
			[]route{
				{method: "GET", path: "/users", target: "https://users.com/"},
				{method: "POST", path: "/x", target: "http://user_service:8081"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			tc.setup(ctx)
			routes := configFromEnv(tc.prefix)

			if want, have := len(tc.expectedRoutes), len(routes); have != want {
				t.Errorf("expected %d routes, found %d", want, have)
			}

			for _, want := range tc.expectedRoutes {
				var found bool
				for _, have := range routes {
					if have == want {
						found = true
					}
				}
				if !found {
					t.Errorf("route not found: %+v", want)
				}
			}
		})
	}
}
