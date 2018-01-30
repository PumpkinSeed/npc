package common

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
)

type Storage struct {
	Name string
	F    *os.File
}

func NewStorage(name string) *Storage {
	return &Storage{
		Name: name,
	}
}

func (s *Storage) init(opt int) error {
	if _, err := os.Stat(s.Name); os.IsNotExist(err) {
		s.F, err = os.Create(s.Name)
		return err
	}

	f, err := os.OpenFile(s.Name, opt, 0644)
	if err != nil {
		return err
	}
	s.F = f
	return nil
}

func (s *Storage) close() error {
	return s.F.Close()
}

func (s *Storage) Write(data ...string) error {
	var err error
	err = s.init(os.O_APPEND | os.O_WRONLY)
	if err != nil {
		return err
	}
	defer s.close()

	for _, v := range data {
		_, err = s.F.WriteString(v + "\n")
	}
	return err
}

func (s *Storage) indexOf(line string) int {
	var counter = 1
	scanner := bufio.NewScanner(s.F)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		if scanner.Text() == line {
			return counter
		}
		counter++
	}
	return -1
}

func (s *Storage) Remove(line string) error {
	var err error
	err = s.init(os.O_APPEND | os.O_RDWR)
	if err != nil {
		return err
	}
	defer s.close()

	var start int
	var n = 1
	start = s.indexOf(line)

	err = s.init(os.O_APPEND | os.O_RDWR)
	if err != nil {
		return err
	}
	defer s.close()

	var b []byte
	if b, err = ioutil.ReadAll(s.F); err != nil {
		return err
	}
	cut, ok := skip(b, start-1)
	if !ok {
		return fmt.Errorf("less than %d lines", start)
	}
	if n == 0 {
		return nil
	}
	tail, ok := skip(cut, n)
	if !ok {
		return fmt.Errorf("less than %d lines after line %d", n, start)
	}
	t := int64(len(b) - len(cut))
	if err = s.F.Truncate(t); err != nil {
		return err
	}
	if len(tail) > 0 {
		_, err = s.F.WriteAt(tail, t)
	}
	return err
}

func skip(b []byte, n int) ([]byte, bool) {
	for ; n > 0; n-- {
		if len(b) == 0 {
			return nil, false
		}
		x := bytes.IndexByte(b, '\n')
		if x < 0 {
			x = len(b)
		} else {
			x++
		}
		b = b[x:]
	}
	return b, true
}
