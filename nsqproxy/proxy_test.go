package nsqproxy_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/pierreprinetti/gateway/nsqproxy"
)

func TestNew(t *testing.T) {
	type checkFunc func(Handler, error) error
	checks := func(fns ...checkFunc) []checkFunc { return fns }

	hasNoError := func(_ Handler, err error) error {
		if err != nil {
			t.Errorf("unexpected error: `%v`", err)
		}
		return nil
	}
	hasError := func(_ Handler, err error) error {
		if err == nil {
			t.Errorf("expected error, found `%v`", err)
		}
		return nil
	}
	hasTopic := func(want string) checkFunc {
		return func(h Handler, _ error) error {
			if h.Topic != want {
				return fmt.Errorf("expected topic %q, found %q", want, h.Topic)
			}
			return nil
		}
	}
	hasAddress := func(want string) checkFunc {
		return func(h Handler, _ error) error {
			if have := h.String(); have != want {
				return fmt.Errorf("expected address %q, found %q", want, have)
			}
			return nil
		}
	}

	testCases := [...]struct {
		name   string
		target string
		checks []checkFunc
	}{
		{
			"parses the host as the address",
			"nsq://host/topic",
			checks(
				hasNoError,
				hasAddress("host"),
			),
		},
		{
			"includes the port in the address",
			"nsq://host:23612/topic",
			checks(
				hasNoError,
				hasAddress("host:23612"),
			),
		},
		{
			"fails on non-url target",
			"�nsq://nsq://invalid",
			checks(
				hasError,
			),
		},
		{
			"parses the path as the topic",
			"nsq://host/topic",
			checks(
				hasNoError,
				hasTopic("topic"),
			),
		},
		{
			"errors on invalid topic",
			"nsq://host/to/pic",
			checks(
				hasError,
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h, err := New(tc.target)

			for _, check := range tc.checks {
				if e := check(h, err); e != nil {
					t.Error(e)
				}
			}
		})
	}
}

type testClient struct {
	returnedError error

	topic string
	msg   []byte
}

func (tc *testClient) Stop()          {}
func (tc *testClient) String() string { return "the-address" }
func (tc *testClient) Publish(topic string, msg []byte) error {
	tc.topic = topic
	tc.msg = msg
	return tc.returnedError
}

func TestHandlerServeHTTP(t *testing.T) {
	type checkFunc func(*httptest.ResponseRecorder, Publisher) error
	checks := func(fns ...checkFunc) []checkFunc { return fns }

	hasStatusCode := func(want int) checkFunc {
		return func(rr *httptest.ResponseRecorder, _ Publisher) error {
			if have := rr.Code; have != want {
				return fmt.Errorf("expected HTTP status code %d, found %d", want, want)
			}
			return nil
		}
	}
	hasPublishedBody := func(want string) checkFunc {
		return func(_ *httptest.ResponseRecorder, p Publisher) error {
			msg := p.(*testClient).msg
			var content struct {
				Url  string `json:"url"`
				Body []byte `json:"body"`
			}
			err := json.Unmarshal(msg, &content)
			if err != nil {
				return fmt.Errorf("unable to read the message: %v", err)
			}
			if have := string(content.Body); have != want {
				return fmt.Errorf("expected message body %q, found %q", want, have)
			}
			return nil
		}
	}
	hasPublishedUrl := func(want string) checkFunc {
		return func(_ *httptest.ResponseRecorder, p Publisher) error {
			msg := p.(*testClient).msg
			var content struct {
				Url  string `json:"url"`
				Body []byte `json:"body"`
			}
			err := json.Unmarshal(msg, &content)
			if err != nil {
				return fmt.Errorf("unable to read the message: %v", err)
			}
			if have := content.Url; have != want {
				return fmt.Errorf("expected message body %q, found %q", want, have)
			}
			return nil
		}
	}
	hasPublishedOnTopic := func(want string) checkFunc {
		return func(_ *httptest.ResponseRecorder, p Publisher) error {
			if have := p.(*testClient).topic; have != want {
				return fmt.Errorf("expected message topic %q, found %q", want, have)
			}
			return nil
		}
	}

	testCases := [...]struct {
		name         string
		requestBody  string
		handlerTopic string
		pubError     error
		checks       []checkFunc
	}{
		{
			name: "returns 201",
			checks: checks(
				hasStatusCode(201),
			),
		},
		{
			name:        "publishes the request body",
			requestBody: "the body",
			checks: checks(
				hasPublishedBody("the body"),
			),
		},
		{
			name: "publishes the request url",
			checks: checks(
				hasPublishedUrl("/send"),
			),
		},
		{
			name:         "publishes on the selected topic",
			handlerTopic: "the-topic",
			checks: checks(
				hasPublishedOnTopic("the-topic"),
			),
		},
		{
			name:     "returns 502 if pub error",
			pubError: http.ErrUnexpectedTrailer,
			checks: checks(
				hasStatusCode(502),
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testC := &testClient{returnedError: tc.pubError}
			h := Handler{
				Publisher: testC,
				Topic:     "the-topic",
			}

			body := strings.NewReader(tc.requestBody)
			req := httptest.NewRequest("POST", "/send", body)
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)

			for _, check := range tc.checks {
				if e := check(rr, h.Publisher); e != nil {
					t.Error(e)
				}
			}
		})
	}
}
