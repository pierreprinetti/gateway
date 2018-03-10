# gateway
[![Build Status](https://travis-ci.org/pierreprinetti/gateway.svg?branch=master)](https://travis-ci.org/pierreprinetti/gateway)
[![Coverage Status](https://coveralls.io/repos/github/pierreprinetti/gateway/badge.svg?branch=master)](https://coveralls.io/github/pierreprinetti/gateway?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/pierreprinetti/gateway)](https://goreportcard.com/report/github.com/pierreprinetti/gateway)

## Configuration

The gateway gathers its configuration from the environment.

### Endpoints

The initialisation function will parse all the environment variable keys starting with the prefix `PROXY_`. The key has to contain the HTTP method and the resource name to be bound to the upstream.

Example:
```shell
export PROXY_GET_users='http://user_service'
export PROXY_PATCH_users='nsq://nsqd:4150/users'
```

With these environment variables:

* every `GET` call to `/users/*` will be routed to `http://user_service`;
* every `PATCH` call to `/users/*` will be forwarded as a NSQ message to the NSQ broker on host `nsqd:4150` with the topic `users`.

### NSQ forwarding

The incoming call is packed in a JSON message containing the url and the body of the request.

Example:
```Javascript
{
  "url": "/users/abc",
  "body": "{\"first_name\":\"John\",\"age\":42}"
}
```
