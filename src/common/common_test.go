package common

import (
	"testing"
)

func TestPseuodUUID(t *testing.T) {
	id, err := pseudoUUID()
	if err != nil {
		t.Error(err)
	}
	t.Log(id)
}

func TestTimestamp(t *testing.T) {
	ti := timestamp()
	t.Log(ti)
}
