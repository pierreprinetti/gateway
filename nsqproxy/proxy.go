// Package nsqproxy translates HTTP calls into NSQ messages.
package nsqproxy

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	nsq "github.com/nsqio/go-nsq"
)

var ErrInvalidTopicName = errors.New("Invalid topic name")

type Publisher interface {
	Publish(topic string, body []byte) error
	Stop()
	String() string
}

type Handler struct {
	Publisher

	Topic string
}

type Message struct {
	Url  string `json:"url"`
	Body []byte `json:"body"`
}

func messageFromRequest(req *http.Request) (*Message, error) {
	u := req.URL.String()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	return &Message{
		Url:  u,
		Body: body,
	}, nil
}

func (h Handler) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	msg, err := messageFromRequest(req)
	if err != nil {
		log.Printf("unable to read the request: %v", err)
		http.Error(wr, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		return
	}

	nsqMessage, err := json.Marshal(msg)
	if err != nil {
		log.Printf("unable to serialise the request data: %v", err)
		http.Error(wr, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		return
	}

	err = h.Publish(h.Topic, nsqMessage)
	if err != nil {
		log.Printf("unable to publish to nsq: %v", err)
		http.Error(wr, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
		return
	}

	wr.WriteHeader(http.StatusCreated)
}

func New(target string) (h Handler, err error) {
	u, err := url.Parse(target)
	if err != nil {
		return h, err
	}

	topic := strings.TrimPrefix(u.Path, "/")

	if !nsq.IsValidTopicName(topic) {
		return h, ErrInvalidTopicName
	}

	prod, err := nsq.NewProducer(u.Host, nsq.NewConfig())
	if err != nil {
		return h, err
	}

	return Handler{
		Publisher: prod,
		Topic:     topic,
	}, nil
}
