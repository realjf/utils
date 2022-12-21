package utils

import (
	"bufio"
	"bytes"
	"io"
	"sync"
)

const (
	STDIN  int = 0
	STDOUT int = 1
	STDERR int = 2
)

type IOReadCloser struct {
	io     io.ReadCloser
	lock   sync.RWMutex
	reader *bufio.Reader
	buf    bytes.Buffer
}

func NewReader(io io.ReadCloser) *IOReadCloser {
	return &IOReadCloser{
		io:     io,
		lock:   sync.RWMutex{},
		reader: bufio.NewReader(io),
		buf:    bytes.Buffer{},
	}
}

func (io *IOReadCloser) Read() (string, error) {
	io.lock.Lock()
	defer io.lock.Unlock()
	str, err := io.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	_, err = io.buf.WriteString(str)
	if err != nil {
		return "", err
	}
	return str, nil
}

func (io *IOReadCloser) Bytes() []byte {
	io.lock.RLock()
	defer io.lock.RUnlock()
	return io.buf.Bytes()
}
