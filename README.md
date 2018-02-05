# npc
RPC framework for NSQ

[![Documentation](https://godoc.org/github.com/PumpkinSeed/npc?status.svg)](https://godoc.org/github.com/PumpkinSeed/npc) 

- [Usage](#usage)
	- [Server](#server)
	- [Client](#client)
	- [Details](#details) (Logger, Configs, AppServer, Interrupter)

## Usage

```
go get github.com/PumpkinSeed/npc
```

### Server

Server implements an exact rpc server which is listening on the request topic and consume the message than publish the answer to the response topic determined in the Envelope (Envelope not seen by the user of the library, it's just for the inner communication).

Create a server using the `npc.New({TYPE})`, `Init({PRODUCER_CONFIG}, {CONSUMER_CONFIG}, {REQUEST_TOPIC}, {NSQ_CHANNEL}, {LOGGER})` and the `Server({rpc.AppServer})`. The type in this case will be `npc.Server`. After start the listener with the `Listen()` method. More information about the [details](#details) section.

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

Client implements a publisher which waiting for the response until it's not arriving or the timeout not reached. 

Create a client using the `npc.New({TYPE})`, `Init({PRODUCER_CONFIG}, {CONSUMER_CONFIG}, {REQUEST_TOPIC}, {NSQ_CHANNEL}, {LOGGER})` and the `Client({RESPONSE_TOPIC})`. The type in this case will be `npc.Client`. After call the `Publish({RESOURCE}, {DATA})` method. More information about the [details](#details) section.

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

### Details

**Logger**

The logger must implement the `common.Logger` interface, by default there is two predefined logger:

- `common.BlankLogger`: Don't log anything
- `common.SimpleLogger`: Log with `fmt.Println`, format the depth of the call and the message

**Producer config**

The producer config supports the following options:

```
type Config struct {
	// NSQConfig config for the consumer
	NSQConfig *nsq.Config

	// NSQDAddress address of the NSQ daemon
	NSQDAddress string

	// Logger of the consumer
	Logger common.Logger

	// LogLevel set the log level of the consumer
	LogLevel nsq.LogLevel
}
```

**Consumer config**

The consumer config supports the following options:

```
type Config struct {
	// NSQConfig config for the consumer
	NSQConfig *nsq.Config

	// NSQDAddress address of the NSQ daemon
	NSQDAddress string

	// NSQLookupdAddresses addresses of the NSQ Lookup daemons
	NSQLookupdAddresses []string

	// Concurrency amout of concurrent handlers of the consumer
	Concurrency int

	// Logger of the consumer
	Logger common.Logger

	// LogLevel set the log level of the consumer
	LogLevel nsq.LogLevel
}
```

**AppServer**

The server of the RPC can have a user-defined server mechanism. It's waiting for an struct which implements the rpc.AppServer interface:

```
type AppServer interface {
	Serve(ctx context.Context, typ string, req []byte) ([]byte, error)
}
```

**Interrupter**

The server have a default interrupter which stops the server in case of `Ctrl+C`, but the user can define an own interrupter and pass it into the server by the following method:

```
SetInterruptor(i func())
```