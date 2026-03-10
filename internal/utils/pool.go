package utils

import "sync"

type Resettable interface {
	Reset()
}

type Pool[T Resettable] struct {
	items []T
	new   func() T
	sync.Mutex
}

func New[T Resettable](newItem func() T) *Pool[T] {
	return &Pool[T]{new: newItem}
}

func (p *Pool[T]) Put(t T) {
	t.Reset()
	p.Mutex.Lock()
	p.items = append(p.items, t)
	p.Mutex.Unlock()
}

func (p *Pool[T]) Get() T {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	
	cnt := len(p.items)
	if cnt == 0 {
		return p.new()
	}
	t := p.items[cnt-1]
	p.items = p.items[:cnt-1]

	return t
}
