# npc
RPC framework for NSQ

[![Documentation](https://godoc.org/github.com/PumpkinSeed/npc?status.svg)](https://godoc.org/github.com/PumpkinSeed/npc) 

## Usage

```
go get github.com/PumpkinSeed/npc
```

### Server

```
package main

import (
	"context"
	"encoding/json"

	"github.com/PumpkinSeed/npc"
	"github.com/PumpkinSeed/npc/lib/common"
	"github.com/PumpkinSeed/npc/lib/consumer"
	"github.com/PumpkinSeed/npc/lib/producer"
	nsq "github.com/nsqio/go-nsq"
)

var localNSQd = "127.0.0.1:4150"

func main() {
	l := common.SingleLogger{}

	pConf := &producer.Config{
		NSQConfig:   nsq.NewConfig(),
		NSQDAddress: localNSQd,
		Logger:      l,
		LogLevel:    nsq.LogLevelInfo,
	}

	cConf := &consumer.Config{
		NSQConfig:   nsq.NewConfig(),
		NSQDAddress: localNSQd,
		Logger:      l,
		LogLevel:    nsq.LogLevelInfo,
	}

	m, err := npc.New(npc.Server).
		Init(pConf, cConf, "request", "server", common.SingleLogger{}).
		Server(&app{})

	if err != nil {
		panic(err)
		return
	}

	// Start to listen as server
	m.Listen()
}

type app struct{}

func (a *app) Serve(ctx context.Context, method string, reqBuf []byte) ([]byte, error) {
	switch method {
	case "Add":
		rsp := struct {
			Z int
		}{
			Z: 12,
		}
		rspBuf, err := json.Marshal(rsp)
		if err != nil {
			return nil, err
		}
		return rspBuf, nil
	default:
		return nil, nil
	}
}
```

### Client

```
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

	rand.Seed(time.Now().UTC().UnixNano())
	action(&wg)
}

func action(wg *sync.WaitGroup) {
	l := common.SingleLogger{}

	pConf := &producer.Config{
		NSQConfig:   nsq.NewConfig(),
		NSQDAddress: localNSQd,
		Logger:      l,
		LogLevel:    nsq.LogLevelInfo,
	}

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
	}

	resp, err := m.Publish("Add", []byte("test"))
	if err != nil {
		panic(err)
	}

	fmt.Println(string(resp))
}
```