package rpc

import (
	"testing"

	nsq "github.com/nsqio/go-nsq"
)

func TestNewClient(t *testing.T) {
	p, _ := nsq.NewProducer("", nsq.NewConfig())
	c := NewClient(p, "request", "response")

	// @todo remove testify
	if c.reqTopic != "request" {
		t.Errorf("reqTopic should be 'request', instead of %s", c.reqTopic)
	}
	if c.rspTopic != "response" {
		t.Errorf("reqTopic should be 'response', instead of %s", c.rspTopic)
	}
	if c.publisher == nil {
		t.Error("publisher should be nil")
	}

	if c.msgNo == 0 {
		t.Error("msgNo should be nil")
	}
}

func TestCorrelationID(t *testing.T) {
	p, _ := nsq.NewProducer("", nsq.NewConfig())
	c := NewClient(p, "request", "response")

	id := c.correlationID()
	t.Log(id)
}

func TestAdd(t *testing.T) {
	rspCh := make(chan *Envelope)
	p, _ := nsq.NewProducer("", nsq.NewConfig())
	c := NewClient(p, "request", "response")
	id := c.correlationID()

	c.add(id, rspCh)
	if c.subscribers[id] == nil {
		t.Error("subscribers[id] should be nil")
	}
}

/*
	The following tests requires local nsq
*/
