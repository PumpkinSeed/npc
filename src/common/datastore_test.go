package common

import (
	"fmt"
	"testing"
)

const dbName = "test.db"

func TestWrite(t *testing.T) {
	s := NewStorage(dbName)
	err := s.Write("lines-123123-123123-123123", "lines-123123-123123-123124", "lines-123123-123123-123125")
	fmt.Println(err)
}

func TestRemove(t *testing.T) {
	s := NewStorage(dbName)
	err := s.Remove("lines-123123-123123-123123")
	fmt.Println(err)
}
