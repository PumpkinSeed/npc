package rpc

import (
	"testing"

	nsq "github.com/nsqio/go-nsq"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	p, _ := nsq.NewProducer("", nsq.NewConfig())
	c := NewClient(p, "request", "response")

	assert.Equal(t, "request", c.reqTopic, "they should be equal")
	assert.Equal(t, "response", c.rspTopic, "they should be equal")

	assert.NotNil(t, c.publisher, "shouldn't be nil")
	assert.NotNil(t, c.msgNo, "shouldn't be nil")
}

/*
	The following tests requires local nsq
*/

func TestCorrelationID(t *testing.T) {
	p, _ := nsq.NewProducer("", nsq.NewConfig())
	c := NewClient(p, "request", "response")

	t.Log(c.msgNo)
	id := c.correlationID()
	t.Log(id)
}
