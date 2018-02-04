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
	// publisher is the nsq producer which responsible
	// for produce the message passed in the Call method
	publisher *nsq.Producer

	// reqTopic stores the request topic name
	reqTopic string

	// rspTopic stores the response topic name
	rspTopic string

	// msgNo determine the id of the message
	// mostly unique and random to identify the
	// the message and it's response
	msgNo uint32

	// subscribers ???
	subscribers map[uint32]chan *Envelope

	// add mutex to handle critical points
	sync.Mutex
}

// NewClient creates new rpc client
// publisher will be used for sending request on reqTopic
// rspTopic will be send in each message envelope, server will reply on that topic
func NewClient(publisher *nsq.Producer, reqTopic, rspTopic string) *Client {
	// rand seed provide a seed mechanism for the msgNo to get a unique identifier
	rand.Seed(time.Now().UnixNano())

	// return the setted up client
	// msgNo gets a random uint number to identify the client's message
	// subscribers creates a channel of Envelopes identified by the msgNo
	return &Client{
		publisher:   publisher,
		reqTopic:    reqTopic,
		rspTopic:    rspTopic,
		msgNo:       rand.Uint32(), //@todo uint64
		subscribers: make(map[uint32]chan *Envelope),
	}
}

// HandleMessage accepts incoming server reponses
// HandleMessage client side handler, if there is a consumed message the
// HandleMessage will handle it and do the necessery steps because of the
// client responsibilities
func (c *Client) HandleMessage(m *nsq.Message) error {
	// fin is the final function called before the function ends
	// the DisableAutoResponse stop go-nsq from responding on that
	// behalf, the Finish the message
	fin := func() {
		m.DisableAutoResponse()
		m.Finish()
	}

	// the Envelope gets decode by the Decode function for further usage
	// if the message body is not an Envelope type of message error will
	// occur and break the execution
	rsp, err := Decode(m.Body)
	if err != nil {
		// raise error without message requeue provided by the fin function
		// ensure the message won't requeue because the invalid body format
		fin()
		return errors.New("envelope unpack failed" + err.Error())
	}

	// find subscriber waiting for response, the list of the subscribers stored
	// and identified with the correlation ID
	if s, found := c.get(rsp.CorrelationID); found {

		// if subscription had been found than pass the response into
		if s != nil {
			s <- rsp
		}
		// when s == nil, means that request timed out, nobody is waiting for response
		// nothing to do in that case
		return nil
	}

	// fin provide the finish of the handle and close the message
	fin()

	// there wasn't subscriber for the message number
	return fmt.Errorf("subscriber not found for %d", rsp.CorrelationID)
}

// Call is the entry-point of the client, it starts the call of the remote function
// the typ determine the resource and the req is the message what it have to send
func (c *Client) Call(ctx context.Context, typ string, req []byte) ([]byte, string, error) {
	return c.CallTopic(ctx, c.reqTopic, typ, req)
}

// CallTopic is the core body of the Call function, it gets the topic to send the
// request and return the exactly same as the Call
func (c *Client) CallTopic(ctx context.Context, reqTopic, typ string, req []byte) ([]byte, string, error) {
	// the request body will be the Envelope encoded version
	// the correlactionID generated based on the msgNo
	correlationID := c.correlationID()
	eReq := &Envelope{
		Method:        typ,
		ReplyTo:       c.rspTopic,
		CorrelationID: correlationID,
		Body:          req,
	}

	// setup deadline, if it's defined in the context
	// the request ExpiresAt ship it to the server as
	// well
	if d, ok := ctx.Deadline(); ok {
		eReq.ExpiresAt = d.Unix()
	}

	// create the channel of the response, it will be a Envelope type
	// add this channel to the list of the sibscribers with the
	// correlationID as the subscriber identifier
	rspCh := make(chan *Envelope)
	c.add(correlationID, rspCh)

	// send request to the server through the nsq publisher
	// defined in the client initializer
	if err := c.publisher.Publish(reqTopic, eReq.Encode()); err != nil {
		return nil, "", errors.New("nsq publish failed" + err.Error())
	}

	// wait for the response in the previously created channel
	// which attached to the subscriber
	// or context timeout/cancelation
	select {
	case rsp := <-rspCh:
		// return the response body and the error of the response,
		// nil as the third error
		return rsp.Body, rsp.Error, nil
	case <-ctx.Done():
		// timeout removes the subscriber
		// returns the context error
		c.timeout(correlationID)
		return nil, "", ctx.Err()
	}
}

// correlationID calculate a new identifier based on the msgNo
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

// add subscriber channel to the list of subscribers
func (c *Client) add(id uint32, ch chan *Envelope) {
	// lock the critical section to avoid race condition
	c.Lock()
	defer c.Unlock()

	// add subscriber channel to the subscribers
	c.subscribers[id] = ch
}

// get the subscriber based on the correlation id attached it to
func (c *Client) get(id uint32) (chan *Envelope, bool) {
	// lock the critical section to avoid race condition
	c.Lock()
	defer c.Unlock()

	// if the subscriber is there it's delete from the list
	// to avoid the memory leaks
	ch, ok := c.subscribers[id]
	if ok {
		delete(c.subscribers, id)
	}
	return ch, ok
}

// timeout do the necessary steps on case of context timeout
func (c *Client) timeout(id uint32) {
	// lock the critical section to avoid race condition
	c.Lock()
	defer c.Unlock()

	// if the subscriber is there it's delete from the list
	// because of the timeout to avoid the memory leaks
	if _, found := c.subscribers[id]; found {
		c.subscribers[id] = nil
	}
}
