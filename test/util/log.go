package util

import (
	"bytes"
	"sync"
)

type SyncBuffer struct {
	b bytes.Buffer
	m sync.Mutex
}

func NewSyncBuffer() *SyncBuffer {
	return &SyncBuffer{b: bytes.Buffer{}}
}

func (b *SyncBuffer) Read(p []byte) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Read(p)
}

func (b *SyncBuffer) Write(p []byte) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Write(p)
}

func (b *SyncBuffer) String() string {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.String()
}
