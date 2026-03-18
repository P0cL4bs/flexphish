package runtime

import (
	"sync"
)

type StepStateStore interface {
	Get(sessionID string) (map[string]interface{}, error)
	Set(sessionID string, values map[string]interface{}) error
	Merge(sessionID string, values map[string]interface{}) error
}

type MemoryStateStore struct {
	data map[string]map[string]interface{}
	mu   sync.RWMutex
}

func NewMemoryStateStore() *MemoryStateStore {
	return &MemoryStateStore{
		data: make(map[string]map[string]interface{}),
	}
}

func (m *MemoryStateStore) Set(sessionID string, values map[string]interface{}) error {

	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[sessionID] = values
	return nil
}

func (m *MemoryStateStore) Merge(sessionID string, values map[string]interface{}) error {

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.data[sessionID]; !ok {
		m.data[sessionID] = map[string]interface{}{}
	}

	for k, v := range values {
		m.data[sessionID][k] = v
	}

	return nil
}

func (m *MemoryStateStore) Get(sessionID string) (map[string]interface{}, error) {

	m.mu.RLock()
	defer m.mu.RUnlock()

	if v, ok := m.data[sessionID]; ok {
		return v, nil
	}

	return map[string]interface{}{}, nil
}
