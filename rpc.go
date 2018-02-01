package rpc

type T int

const (
	Server T = iota
	Client
)

type Main struct {
	Server bool
	Client bool
}

func New(t T) Main {
	m := Main{}
	switch t {
	case Server:
		m.Server = true
	case Client:
		m.Client = true
	}

	return m
}
