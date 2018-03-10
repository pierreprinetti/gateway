# gateway
[![Build Status](https://travis-ci.org/pierreprinetti/gateway.svg?branch=master)](https://travis-ci.org/pierreprinetti/gateway)
[![Coverage Status](https://coveralls.io/repos/github/pierreprinetti/gateway/badge.svg?branch=master)](https://coveralls.io/github/pierreprinetti/gateway?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/pierreprinetti/gateway)](https://goreportcard.com/report/github.com/pierreprinetti/gateway)

This is a simplistic reverse proxy for HTTP and NSQ upstreams.

It gathers its configuration from the environment to map a resouce with an upstream. Every HTTP verb has to be configured independently.

## Configuration

The initialisation function will parse all the environment variable keys starting with the prefix `PROXY_`. The key has to contain the HTTP method and the resource name to be bound to the upstream.

Example:
```shell
export PROXY_GET_users='http://user_service'
export PROXY_PATCH_users='nsq://nsqd:4150/users'
```

With these environment variables:

* every `GET` call to `/users/*` will be routed to `http://user_service`;
* every `PATCH` call to `/users/*` will be forwarded as a NSQ message to the NSQ broker on host `nsqd:4150` with the topic `users`.

### In a docker-compose.yml

This gateway is available in the public Docker registry as `pierreprinetti/gateway`.

```yaml
version: '3'
services:
  gateway:
    image: pierreprinetti:gateway
    environment:
      PROXY_GET_users: 'http://user_service'
      PROXY_PATCH_users: 'nsq://nsqd:4150/users'
    depends_on:
      - user_service
      - nsqd
    ports:
      - '80:80'
```

### NSQ forwarding

The nsq connection string must have `nsq` as the protocol and be in the form:

```
nsq://host:port/topic
```

The incoming HTTP requests are packed in a JSON message containing the url and the body of the request.

Example:
```Javascript
{
  "url": "/users/abc",
  "body": "{\"first_name\":\"John\",\"age\":42}"
}
```
