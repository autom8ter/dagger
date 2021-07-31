package driver

import (
	"github.com/autom8ter/dagger/constants"
	"sort"
	"sync"
)

type inMemIndex struct {
	cacheMap   map[string]map[string]interface{}
	namespaces map[string]struct{}
	mu         sync.RWMutex
	closeOnce  sync.Once
}

func (n *inMemIndex) Len(namespace string) int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	if c, ok := n.cacheMap[namespace]; ok {
		return len(c)
	}
	return 0
}

func (n *inMemIndex) Namespaces() []string {
	var namespaces []string
	n.mu.RLock()
	defer n.mu.RUnlock()
	for k, _ := range n.namespaces {
		namespaces = append(namespaces, k)
	}
	sort.Strings(namespaces)
	return namespaces
}

func (n *inMemIndex) Get(namespace string, key string) (interface{}, bool) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	if c, ok := n.cacheMap[namespace]; ok {
		return c[key], true
	}
	return nil, false
}

func (n *inMemIndex) Set(namespace string, key string, value interface{}) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.cacheMap[namespace] == nil {
		n.cacheMap[namespace] = map[string]interface{}{}
	}
	n.namespaces[namespace] = struct{}{}
	n.cacheMap[namespace][key] = value
}

func (n *inMemIndex) Range(namespace string, f func(key string, value interface{}) bool) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	if namespace == constants.AnyType {
		for _, c := range n.cacheMap {
			for k, v := range c {
				f(k, v)
			}
		}
	} else {
		if c, ok := n.cacheMap[namespace]; ok {
			for k, v := range c {
				f(k, v)
			}
		}
	}
}

func (n *inMemIndex) Delete(namespace string, key string) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	if c, ok := n.cacheMap[namespace]; ok {
		delete(c, key)
	}
}

func (n *inMemIndex) Exists(namespace string, key string) bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	_, ok := n.cacheMap[namespace][key]
	return ok
}

func (n *inMemIndex) Clear(namespace string) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	if cache, ok := n.cacheMap[namespace]; ok {
		for k, _ := range cache {
			delete(cache, k)
		}
	}
}

func (n *inMemIndex) Close() {
	n.closeOnce.Do(func() {
		n.mu.Lock()
		defer n.mu.Unlock()
		for namespace, _ := range n.cacheMap {
			n.Clear(namespace)
		}
	})
}
