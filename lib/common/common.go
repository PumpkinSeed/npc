package common

import (
	"fmt"
)

type Discoverer interface {
	NSQDAddress() (string, error)
	NSQLookupdAddresses() ([]string, error)
	Subscribe(Subscriber)
	NodeName() string
}

type Logger interface {
	Output(calldepth int, s string) error
}

type Subscriber interface {
	DisconnectFromNSQLookupd(addr string) error
	ConnectToNSQLookupd(addr string) error
}

type BlankLogger struct{}

func (BlankLogger) Output(calldepth int, s string) error {
	return nil
}

type SingleLogger struct{}

func (SingleLogger) Output(calldepth int, s string) error {
	fmt.Printf("depth: %d, msg: %s\n", calldepth, s)
	return nil
}
