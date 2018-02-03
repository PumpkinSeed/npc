package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/PumpkinSeed/npc"
	"github.com/PumpkinSeed/npc/lib/common"
	"github.com/PumpkinSeed/npc/lib/consumer"
	"github.com/PumpkinSeed/npc/lib/producer"
	nsq "github.com/nsqio/go-nsq"
)

const (
	reqTopic = "request"  // topic for sending request to server
	rspTopic = "response" // topic for getting responses from server
	channel  = "client"   // channel name for rspTopic topic
)

var localNSQd = "127.0.0.1:4150"

func main() {
	var wg sync.WaitGroup
	//defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()
	rand.Seed(time.Now().UTC().UnixNano())
	for i := 0; i < 100; i++ {
		//wg.Add(1)
		action(&wg)
		//	fmt.Println(runtime.NumGoroutine())
	}
	//wg.Wait()
}

func action(wg *sync.WaitGroup) {
	l := common.SingleLogger{}

	pConf := &producer.Config{
		NSQConfig:   nsq.NewConfig(),
		NSQDAddress: localNSQd,
		Logger:      l,
		LogLevel:    nsq.LogLevelInfo,
	}

	// create consumer arround rpcClient
	cConf := &consumer.Config{
		NSQConfig:   nsq.NewConfig(),
		NSQDAddress: localNSQd,
		Logger:      l,
		LogLevel:    nsq.LogLevelInfo,
	}

	m, err := npc.New(npc.Client).
		Init(pConf, cConf, "request", "server", l).
		Client("response")

	if err != nil {
		panic(err)
		return
	}

	resp, err := m.Publish("Add", []byte("test"))
	if err != nil {
		panic(err)
		return
	}

	fmt.Println(string(resp))
}
