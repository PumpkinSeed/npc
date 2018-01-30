package main

import (
	"flag"
	"math/rand"
	"time"

	"github.com/PumpkinSeed/durable-nsq-example/src/common"
	"github.com/PumpkinSeed/durable-nsq-example/src/consumer"
	"github.com/PumpkinSeed/durable-nsq-example/src/producer"
)

const (
	topic = "lines"
	ch    = "metrics"
	log   = true
)

var consume bool
var amount int

func init() {
	rand.Seed(time.Now().UnixNano())
	flag.BoolVar(&consume, "consume", false, "Consume the queue")
	flag.IntVar(&amount, "amount", 10, "Amount of the test messages")
	flag.Parse()
}

func main() {
	if consume {
		go func() {
			for i := 0; i < amount; i++ {
				produceHandler(randStringRunes(10))
			}
		}()
		consumeHandler()
	} else {
		for i := 0; i < amount; i++ {
			produceHandler(randStringRunes(10))
		}
	}
}

func consumeHandler() {
	config := consumer.NewConfig(topic, ch, log, &consumer.DefaultMessageHandler{})
	consumer.Start(config)

	// 1. Start to consume
	// 2. Message Handler produce back to the lines+id topic
}

func produceHandler(message string) error {
	config := producer.NewConfig(topic, log)

	m, err := common.NewMessage(message)
	if err != nil {
		return err
	}
	data, err := common.Encode(m)
	if err != nil {
		return err
	}
	producer.Write(data, config)

	// 3. Start to wait for reply in lines+{id}
	return nil
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	t := time.Now()
	return t.Format("2006-01-02 15:04:05") + " - " + string(b)
}
