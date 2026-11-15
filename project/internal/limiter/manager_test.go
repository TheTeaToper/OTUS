package limiter

import (
	"testing"
	"time"
)

func TestManager_CleanupExpiredBuckets(t *testing.T) {
	ttl := 100 * time.Millisecond
	m := NewManager(ttl)
	defer m.Stop()

	// Создаем бакет и делаем запрос
	b1 := m.GetBucket("user1", 10)
	b1.Allow()

	time.Sleep(150 * time.Millisecond)
	m.cleanup()

	m.mutex.RLock()
	_, exists := m.buckets["user1"]
	m.mutex.RUnlock()

	if exists {
		t.Fatal("Неактивный бакет должен был удалиться после истечения TTL")
	}
}
