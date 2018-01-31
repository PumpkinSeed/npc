package common

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
