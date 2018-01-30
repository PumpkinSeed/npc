package utils

import (
	"math/rand"
	"time"
)

func GetTopicName(topic, id string) string {
	return topic + "-" + id
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	t := time.Now()
	return t.Format("2006-01-02 15:04:05") + " - " + string(b)
}
