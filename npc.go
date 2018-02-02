package npc

// T the type of the npc
type T int

const (
	// Server type
	Server T = iota

	// Client type
	Client
)

// Main handler
type Main struct {
	Server bool
	Client bool
}

// New creates a new instance of the Main handler based on the type
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
