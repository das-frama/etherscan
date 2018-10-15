package main

import (
	"sync"
)

type Holder struct {
	Address     string
	Balance     string
	Transcation int
	LastActive  int64
}

type Holders struct {
	mx sync.RWMutex
	m  map[string]*Holder
}

func (h *Holders) Get(key string) (*Holder, bool) {
	h.mx.RLock()
	defer h.mx.RUnlock()
	val, ok := h.m[key]
	return val, ok
}

func (h *Holders) Set(key string, value *Holder) {
	h.mx.Lock()
	defer h.mx.Unlock()
	h.m[key] = value
}

func NewHolders() *Holders {
	return &Holders{
		m: make(map[string]*Holder),
	}
}
