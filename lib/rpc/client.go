package rpc

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/nsqio/go-nsq"
)

// Client rpc client side
type Client struct {
	publisher   *nsq.Producer
	reqTopic    string
	rspTopic    string
	msgNo       uint32
	subscribers map[uint32]chan *Envelope
	sync.Mutex
}

// NewClient creates new rpc client.
// publisher will be used for sending request on reqTopic.
// rspTopic will be send in each message envelope, server will reply on that topic.
func NewClient(publisher *nsq.Producer, reqTopic, rspTopic string) *Client {
	// seed the msgNo
	rand.Seed(time.Now().UnixNano())

	logg("#-NewClient: client init")
	// return the setted up client
	return &Client{
		publisher:   publisher,
		reqTopic:    reqTopic,
		rspTopic:    rspTopic,
		msgNo:       rand.Uint32(), //@todo uint64
		subscribers: make(map[uint32]chan *Envelope),
	}
}

// HandleMessage accepts incoming server reponses.
func (c *Client) HandleMessage(m *nsq.Message) error {
	logg("#-Client-HandleMessage: fin define")
	fin := func() {
		m.DisableAutoResponse()
		m.Finish()
	}

	logg("#-Client-HandleMessage: decode body")
	// unpack message
	rsp, err := Decode(m.Body)
	if err != nil {
		fin()
		logg("#-Client-CallTopic: " + err.Error())
		return errors.New("envelope unpack failed" + err.Error())
	}

	logg("#-Client-HandleMessage: find subscriber")
	// find subscriber waiting for response
	if s, found := c.get(rsp.CorrelationID); found {
		if s != nil {
			s <- rsp
		}
		// when s == nil, means that request timed out, nobody is waiting for response
		// nothing to do in that case
		return nil
	}

	logg("#-Client-HandleMessage: fin")
	fin()
	return fmt.Errorf("subscriber not found for %d", rsp.CorrelationID)
}

func (c *Client) Call(ctx context.Context, typ string, req []byte) ([]byte, string, error) {
	logg("#-Client-Call: CallTopic")
	return c.CallTopic(ctx, c.reqTopic, typ, req)
}

// Call entry point for request from application.
func (c *Client) CallTopic(ctx context.Context, reqTopic, typ string, req []byte) ([]byte, string, error) {
	// craete envelope
	logg("#-Client-CallTopic: correlationID")
	correlationID := c.correlationID()
	eReq := &Envelope{
		Method:        typ,
		ReplyTo:       c.rspTopic,
		CorrelationID: correlationID,
		Body:          req,
	}
	logg("#-Client-CallTopic: deadline")
	if d, ok := ctx.Deadline(); ok {
		eReq.ExpiresAt = d.Unix()
	}
	rspCh := make(chan *Envelope)
	// subscriebe for response on that correlationID
	logg("#-Client-CallTopic: add")
	c.add(correlationID, rspCh)
	// send request to the server

	logg("#-Client-CallTopic: publish")
	if err := c.publisher.Publish(reqTopic, eReq.Encode()); err != nil {
		logg("#-Client-CallTopic: " + err.Error())
		return nil, "", errors.New("nsq publish failed" + err.Error())
	}
	// wiat for response or context timeout/cancelation
	logg("#-Client-CallTopic: wait for response")
	select {
	case rsp := <-rspCh:
		return rsp.Body, rsp.Error, nil
	case <-ctx.Done():
		logg("#-Client-CallTopic: timeout")
		c.timeout(correlationID)
		return nil, "", ctx.Err()
	}
}

// get correlation id
func (c *Client) correlationID() uint32 {
	// lock the critical section to avoid race condition
	c.Lock()
	defer c.Unlock()

	// if the msgNo is the max of the uint32 set to 0
	if c.msgNo == math.MaxUint32 {
		c.msgNo = 0
	} else {
		c.msgNo++
	}
	return c.msgNo
}

// add subscriber channel to the subsrcibers
func (c *Client) add(id uint32, ch chan *Envelope) {
	// lock the critical section to avoid race condition
	c.Lock()
	defer c.Unlock()

	// add subscriber channel to the subsrcibers
	c.subscribers[id] = ch
}

func (c *Client) get(id uint32) (chan *Envelope, bool) {
	c.Lock()
	defer c.Unlock()
	ch, ok := c.subscribers[id]
	if ok {
		delete(c.subscribers, id)
	}
	return ch, ok
}

func (c *Client) timeout(id uint32) {
	c.Lock()
	defer c.Unlock()
	if _, found := c.subscribers[id]; found {
		c.subscribers[id] = nil
	}
}
