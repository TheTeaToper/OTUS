package limiter

import (
	"sync"
	"time"
)

type Manager struct {
	mutex   sync.RWMutex
	buckets map[string]*Bucket
	// Время жизни неактивного бакета (например, 5-10 минут).
	ttl         time.Duration
	stopChannel chan struct{}
}

func NewManager(ttl time.Duration) *Manager {
	manager := &Manager{
		buckets:     make(map[string]*Bucket),
		ttl:         ttl,
		stopChannel: make(chan struct{}),
	}

	if ttl > 0 {
		go manager.cleanupLoop()
	}

	return manager
}

func (m *Manager) GetBucket(key string, limit int) *Bucket {
	m.mutex.RLock()
	bucket, exists := m.buckets[key]
	m.mutex.RUnlock()

	if exists {
		return bucket
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Двойная проверка на случай race condition.
	if bucket, exists = m.buckets[key]; exists {
		return bucket
	}

	bucket = NewBucket(limit)
	m.buckets[key] = bucket
	return bucket
}

// ResetKey удаляет/сбрасывает бакеты для конкретного ключа.
func (m *Manager) ResetKey(key string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.buckets, key)
}

// Останов фоновой очистки неактивных бакетов.
func (m *Manager) Stop() {
	select {
	case <-m.stopChannel:
	default:
		close(m.stopChannel)
	}
}

func (m *Manager) cleanupLoop() {
	// Проверка неактивных бакетов с интервалом половины TTL (но не реже 1 раза в секунду).
	interval := m.ttl / 2
	if interval < time.Second {
		interval = time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.cleanup()
		case <-m.stopChannel:
			return
		}
	}
}

func (m *Manager) cleanup() {
	now := time.Now()

	m.mutex.Lock()
	defer m.mutex.Unlock()

	for key, bucket := range m.buckets {
		if now.Sub(bucket.LastAccess()) > m.ttl {
			delete(m.buckets, key)
		}
	}
}
