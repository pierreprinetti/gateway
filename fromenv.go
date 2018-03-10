package main

import (
	"os"
	"strings"
)

// routeFromEnvVar returns a route from an environment variable
// beginning with prefix.
// env is expected in the form "prefixkey=value".
// ok is true if a route has been successfully parsed out of env.
func routeFromEnvVar(prefix, env string) (r route, ok bool) {
	envKV := strings.SplitN(env, "=", 2)
	if len(envKV) < 2 || !strings.HasPrefix(envKV[0], prefix) {
		return
	}

	if envKV[1] == "" {
		return
	}

	envKV[0] = envKV[0][len(prefix):]

	methodResource := strings.SplitN(envKV[0], "_", 2)
	if len(methodResource) < 2 {
		return
	}

	r.method = methodResource[0]
	r.path = "/" + methodResource[1]
	r.target = envKV[1]

	return r, true
}

// configFromEnv parses the environment and returns a slice of routes
// successfully parsed.
func configFromEnv(prefix string) []route {
	var routes []route

	for _, env := range os.Environ() {
		if r, ok := routeFromEnvVar(prefix, env); ok {
			routes = append(routes, r)
		}
	}
	return routes
}
