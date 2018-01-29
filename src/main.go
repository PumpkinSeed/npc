package main

import (
	"flag"
	"math/rand"
	"time"

	"github.com/PumpkinSeed/durable-nsq-example/src/consumer"
	"github.com/PumpkinSeed/durable-nsq-example/src/producer"
)

var consume bool
var amount int

func init() {
	rand.Seed(time.Now().UnixNano())
	flag.BoolVar(&consume, "consume", false, "Consume the queue")
	flag.IntVar(&amount, "amount", 10, "Amount of the test messages")
	flag.Parse()
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStringRunes(n int) []byte {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	t := time.Now()
	return []byte(t.Format("2006-01-02 15:04:05") + " - " + string(b))
}

func main() {
	if consume {
		go func() {
			for i := 0; i < amount; i++ {
				producer.Write(randStringRunes(10))
			}
		}()
		consumer.Start()
	} else {
		for i := 0; i < amount; i++ {
			producer.Write(randStringRunes(10))
		}
	}
}
