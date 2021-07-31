package driver

import "sync"

type Index interface {
	Len(namespace string) int
	Namespaces() []string
	Get(namespace string, key string) (interface{}, bool)
	Set(namespace string, key string, value interface{})
	Range(namespace string, f func(key string, value interface{}) bool)
	Delete(namespace string, key string)
	Exists(namespace string, key string) bool
	Clear(namespace string)
	Close()
}

func NewInMemIndex() Index {
	return &inMemIndex{
		cacheMap:   map[string]map[string]interface{}{},
		mu:         sync.RWMutex{},
		closeOnce:  sync.Once{},
		namespaces: map[string]struct{}{},
	}
}

type Queue interface {
	Enqueue(val interface{})
	Dequeue() (interface{}, bool)
	IsEmpty() bool
	Len() int
	Range(fn func(element interface{}) bool)
}

func NewInMemQueue() Queue {
	vals := &queue{[]interface{}{}}
	return vals
}

type Stack interface {
	Pop() (interface{}, bool)
	Push(f interface{})
	Range(fn func(element interface{}) bool)
	IsEmpty() bool
	Len() int
}

func NewInMemStack() Stack {
	vals := stackI([]interface{}{})
	return &vals
}