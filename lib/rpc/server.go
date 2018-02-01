package rpc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nsqio/go-nsq"
)

var (
	requeueDelay = time.Second

	// touchInterval for long processing messages
	// resets nsqd timeout for in-flight message
	// good for default nsqd timeout (1m)
	touchInterval = 45 * time.Second
)

type appServer interface {
	Serve(ctx context.Context, typ string, req []byte) ([]byte, error)
}

// Server rpc server side.
type Server struct {
	ctx      context.Context
	srv      appServer
	producer *nsq.Producer
}

// NewServer creates new rpc server for appServer.
// producer will be used for sending replies.
func NewServer(ctx context.Context, srv appServer, producer *nsq.Producer) *Server {
	return &Server{
		ctx:      ctx,
		srv:      srv,
		producer: producer,
	}
}

// HandleMessage server side handler.
func (s *Server) HandleMessage(m *nsq.Message) error {
	fin := func() {
		m.DisableAutoResponse()
		m.Finish()
	}
	// decode message
	req, err := Decode(m.Body)
	if err != nil {
		fin() // raise error without message requeue
		return errors.New("envelope unpack failed: " + err.Error())
	}
	// check expiration
	if req.Expired() {
		fin()
		return fmt.Errorf("expired %s %d", req.Method, req.CorrelationID)
	}
	// periodically call touch on the nsq message while app is still processing it
	defer touchMessage(s.ctx, m)()
	// call aplication
	appRsp, appErr := s.srv.Serve(s.ctx, req.Method, req.Body)
	if s.ctx.Err() != nil || appErr == context.Canceled {
		// context timeout/cancel
		// notice that we are also requeuing on appErr == context.Cancel
		// that's mechanism for application to postpone processing of the message
		m.RequeueWithoutBackoff(requeueDelay)
		return nil
	}
	// need to reply
	if req.ReplyTo == "" {
		return nil
	}
	// create reply
	rsp := req.Reply(appRsp, appErr)
	// send reply
	if err := s.producer.Publish(req.ReplyTo, rsp.Encode()); err != nil {
		return errors.New("nsq publish failed: " + err.Error())
	}
	return nil
}

// touchMessage to prevent auto-requeing in the nsqd
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
