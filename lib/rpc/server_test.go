package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/PumpkinSeed/nsq-rpc/lib/common"
	"github.com/PumpkinSeed/nsq-rpc/lib/consumer"
	"github.com/PumpkinSeed/nsq-rpc/lib/producer"
	nsq "github.com/nsqio/go-nsq"
)

/*
	The following tests requires local nsqd
*/

const (
	reqTopic = "request"  // topic for listening to requests
	rspTopic = "response" // topic for getting responses from server
	channel  = "server"   // channel name on that topic
)

var localNSQd = "127.0.0.1:4150"

var wg sync.WaitGroup

func TestServer(t *testing.T) {
	wg.Add(1)
	go s(t)

	c(t)
}

func c(t *testing.T) {
	l := common.SingleLogger{}

	pConf := &producer.Config{
		NSQConfig:   nsq.NewConfig(),
		NSQDAddress: localNSQd,
		Logger:      l,
		LogLevel:    nsq.LogLevelInfo,
	}

	p, err := producer.New(pConf)

	// rpc client: sends requests, waits and accepts responses
	//             provides interface for application
	rpcClient := NewClient(p, reqTopic, rspTopic)

	// create consumer arround rpcClient
	cConf := &consumer.Config{
		NSQConfig:   nsq.NewConfig(),
		NSQDAddress: localNSQd,
		Logger:      l,
		LogLevel:    nsq.LogLevelInfo,
	}
	c, err := consumer.New(cConf, rspTopic, channel, rpcClient)
	if err != nil {
		log.Fatal(err)
	}

	// application client
	client := &client{t: rpcClient}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	// clean exit
	defer p.Stop() // 3. stop producing new requests
	defer cancel() // 2. cancel any pending (waiting for responses)
	defer c.Stop() // 1. stop listening for responses
	x := randInt(1, 30)
	y := randInt(1, 30)
	z, err := client.Add(ctx, x, y)
	if err != nil {
		log.Fatal(err)
	}
	wg.Wait()
	fmt.Printf("=> %d + %d =  %d\n", x, y, z)
}

func s(t *testing.T) {
	l := common.SingleLogger{}

	pConf := &producer.Config{
		NSQConfig:   nsq.NewConfig(),
		NSQDAddress: localNSQd,
		Logger:      l,
		LogLevel:    nsq.LogLevelInfo,
	}

	p, err := producer.New(pConf)

	// rpc server: accepts request, calls application, sends response
	ctx, cancel := context.WithCancel(context.Background())
	appServer := &server{}
	rpcServer := NewServer(ctx, appServer, p)

	// consumer arround rpcServer
	cConf := &consumer.Config{
		NSQConfig:   nsq.NewConfig(),
		NSQDAddress: localNSQd,
		Logger:      l,
		LogLevel:    nsq.LogLevelInfo,
	}
	c, err := consumer.New(cConf, reqTopic, channel, rpcServer)
	if err != nil {
		log.Fatal(err)
	}

	// clean exit
	defer p.Stop() // 3. stop response producer
	defer cancel() // 2. cancel any pending operation (returns unfinished messages to nsq)
	defer c.Stop() // 1. stop accepting new requests

}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

// transport is application interface for sending request to the remote server
// method - server method name
// req    - request
// returns:
//   reponse
//   application error, string, "" if there is no error
//   transport error
type transport interface {
	Call(ctx context.Context, method string, req []byte) ([]byte, string, error)
}

type client struct {
	t transport
}

func (c *client) Add(ctx context.Context, x, y int) (int, error) {
	defer wg.Done()
	req := &request{X: x, Y: y}
	// marshall request
	reqBuf, err := json.Marshal(req)
	if err != nil {
		return 0, err
	}
	// pass request to trasport, and get response
	rspBuf, _, err := c.t.Call(ctx, "Add", reqBuf)
	// request was canceled
	if ctx.Err() != nil {
		return 0, ctx.Err()
	}
	// request failed
	if err != nil {
		return 0, err
	}
	// unmarshal response
	var rsp response
	err = json.Unmarshal(rspBuf, &rsp)
	if err != nil {
		return 0, err
	}
	return rsp.Z, nil
}

// dto structures

type request struct {
	X int
	Y int
}

type response struct {
	Z int
}

func waitForInterupt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
}

type server struct{}

// Server is entry point for all rpc requests
// method is the name of the server method
// reqBuf request data
// returns:
//   response data
//   application error
func (s *server) Serve(ctx context.Context, method string, reqBuf []byte) ([]byte, error) {
	switch method {
	case "Add":
		// unpack
		var req request
		err := json.Unmarshal(reqBuf, &req)
		if err != nil {
			return nil, err
		}
		// call actual server method
		z := s.add(req.X, req.Y)
		// pack
		rsp := response{Z: z}
		rspBuf, err := json.Marshal(rsp)
		if err != nil {
			return nil, err
		}
		return rspBuf, nil
	default:
		return nil, nil
	}
}

// actual server method
func (s *server) add(x, y int) int {
	return x + y
}
