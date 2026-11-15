package limiter

import (
	"sync"
	"time"
)

type Bucket struct {
	mutex sync.Mutex
	// Сколько токенов добавляется в секунду (Лимит / 60)
	rate float64
	// Максимальное количество токенов (равно N, M или K)
	capacity float64
	// Текущее количество доступных токенов
	tokens float64
	// Время последнего обновления токенов
	lastRefreshed time.Time
}

func NewBucket(limitPerMinute int) *Bucket {
	return &Bucket{
		rate:          float64(limitPerMinute) / 60.0,
		capacity:      float64(limitPerMinute),
		tokens:        float64(limitPerMinute),
		lastRefreshed: time.Now(),
	}
}

func (b *Bucket) Allow() bool {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastRefreshed).Seconds()

	if elapsed > 0 {
		b.tokens += elapsed * b.rate
		if b.tokens > b.capacity {
			b.tokens = b.capacity
		}
		b.lastRefreshed = now
	}

	if b.tokens >= 1.0 {
		b.tokens -= 1.0
		return true
	}
	return false
}

func (b *Bucket) Reset() {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.tokens = b.capacity
	b.lastRefreshed = time.Now()
}

func (b *Bucket) LastAccess() time.Time {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	return b.lastRefreshed
}
