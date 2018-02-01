package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/PumpkinSeed/nsq-rpc/lib/common"
	"github.com/PumpkinSeed/nsq-rpc/lib/consumer"
	"github.com/PumpkinSeed/nsq-rpc/lib/producer"
	"github.com/PumpkinSeed/nsq-rpc/lib/rpc"
	nsq "github.com/nsqio/go-nsq"
)

const (
	reqTopic = "request"  // topic for sending request to server
	rspTopic = "response" // topic for getting responses from server
	channel  = "client"   // channel name for rspTopic topic
)

var localNSQd = "127.0.0.1:4150"

func main() {
	//var wg sync.WaitGroup
	//defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()
	rand.Seed(time.Now().UTC().UnixNano())
	//for i := 0; i < 100; i++ {
	//	wg.Add(1)
	action()
	//	fmt.Println(runtime.NumGoroutine())
	//}
	//wg.Wait()
}

func action() {

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
	rpcClient := rpc.NewClient(p, reqTopic, rspTopic)

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
	fmt.Printf("%d + %d =  %d\n", x, y, z)
	// defer wg.Done()
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
