package rpc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nsqio/go-nsq"
)

var (
	// determine the delay what the server use to
	// requeue the message without backoff
	requeueDelay = time.Second

	// touchInterval for long processing messages
	// resets nsqd timeout for in-flight message
	// good for default nsqd timeout (1m)
	touchInterval = 45 * time.Second
)

// appServer determine the entry point for all rpc request
// typ is the resource what the client want to reach
// req is the body of the request
type appServer interface {
	Serve(ctx context.Context, typ string, req []byte) ([]byte, error)
}

// Server is the exact rpc server handle the context
// srv stores the entry point of the server
// producer is an nsq producer based on the replyTo topic
// which is publish the response
type Server struct {
	ctx      context.Context
	srv      appServer
	producer *nsq.Producer
}

// NewServer creates new rpc server for appServer
// producer will be used for sending replies
func NewServer(ctx context.Context, srv appServer, producer *nsq.Producer) *Server {
	return &Server{
		ctx:      ctx,
		srv:      srv,
		producer: producer,
	}
}

// HandleMessage server side handler, if there is a consumed message the
// HandleMessage will handle it and do the necessery steps because of the
// server responsibilities it will publish the response as well, returned
// by the appServer
func (s *Server) HandleMessage(m *nsq.Message) error {
	// fin is the final function called before the function ends
	// ..... @todo what is ...?
	fin := func() {
		m.DisableAutoResponse()
		m.Finish()
	}

	// the Envelope gets decode by the Decode function for further usage
	// if the message body is not an Envelope type of message error will
	// occur and break the execution
	req, err := Decode(m.Body)
	if err != nil {
		// raise error without message requeue provided by the fin function
		// ensure the message won't requeue because the invalid body format
		fin()
		return errors.New("envelope unpack failed: " + err.Error())
	}

	// check expiration, if it's expired it's also provide an error, because
	// in this case the client is no longer waiting for the answer
	if req.Expired() {
		fin() // @todo figure out to requeue
		return fmt.Errorf("expired %s %d", req.Method, req.CorrelationID)
	}

	// periodically call touch on the nsq message while app is still processing it
	// issue what it solve in long-running consumers provided at the function definition
	defer touchMessage(s.ctx, m)()

	// call the user defined entry point to get the response of the request
	// provide err and response from the appServer
	appRsp, appErr := s.srv.Serve(s.ctx, req.Method, req.Body)

	// context timeout/cancel
	// notice that we are also requeuing on appErr == context.Cancel
	// that's mechanism for application to postpone processing of the message
	if s.ctx.Err() != nil || appErr == context.Canceled {
		m.RequeueWithoutBackoff(requeueDelay)
		return nil
	}

	// need to reply so if the replyTo is empty it will keep on hold the client
	// because there isn't reply topic
	if req.ReplyTo == "" {
		return nil
	}

	// the Envelope has a Reply method which creates the response of the rpc call
	rsp := req.Reply(appRsp, appErr)

	// the producer of the server sends the Envelope response to the reply topic
	// the producer defined on the initial level, passed as a pointer for the
	// reusability, and memory safe workflow
	if err := s.producer.Publish(req.ReplyTo, rsp.Encode()); err != nil {
		return errors.New("nsq publish failed: " + err.Error())
	}
	return nil
}

// touchMessage to prevent auto-requeing in the nsqd
// https://github.com/nsqio/go-nsq/issues/204
func touchMessage(ctx context.Context, m *nsq.Message) func() {
	ctxTouch, cancel := context.WithCancel(ctx)
	go func() {
		for {
			select {
			case <-ctxTouch.Done():
				return
			case <-time.After(touchInterval):
				m.Touch()
			}
		}
	}()
	return cancel
}
