package syncmap

import "sync"

type Map[K comparable, V any] struct {
	m   map[K]V
	mux sync.RWMutex
}

func (m *Map[K, V]) Get(k K) (V, bool) {
	m.mux.RLock()
	defer m.mux.RUnlock()
	v, ok := m.m[k]
	return v, ok
}

func (m *Map[K, V]) Put(k K, v V) {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.m[k] = v
}

func (m *Map[K, V]) Delete(k K) {
	m.mux.Lock()
	defer m.mux.Unlock()
	// delete the key from the map
	if _, ok := m.m[k]; ok {
		delete(m.m, k)
	}
}

func (m *Map[K, V]) Range(f func(key K, value V) bool) {
	for k, v := range m.m {
		// call the function with the key and value
		if !f(k, v) {
			return // stop iteration if f returns false
		}
	}
}
func (m *Map[K, V]) Values() []V {
	m.mux.RLock()
	defer m.mux.RUnlock()
	values := make([]V, 0, len(m.m))
	for _, v := range m.m {
		values = append(values, v)
	}
	return values
}

func NewMap[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{m: make(map[K]V)}

}
