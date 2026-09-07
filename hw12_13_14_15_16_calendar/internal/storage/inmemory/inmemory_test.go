package inmemorystorage

import (
	"sync"
	"testing"
	"time"

	"github.com/TheTeaToper/OTUS/hw12_13_14_15_16_calendar/internal/storage"
	// inmemorystorage "github.com/TheTeaToper/OTUS/hw12_13_14_15_16_calendar/internal/storage/inmemory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Основная логика (CRUD)
func TestInmemoryStorage_CRUD(t *testing.T) {
	store := New()

	now := time.Now()
	event := &storage.Event{
		Title:     "First event",
		StartTime: now,
		EndTime:   now.Add(1 * time.Hour),
	}

	// Add
	err := store.Add(event)
	require.NoError(t, err)
	assert.Equal(t, int64(1), event.ID)

	// List
	list, err := store.List()
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "First event", list[0].Title)

	// Update
	event.Title = "Updated event"
	err = store.Update(event)
	require.NoError(t, err)

	list, err = store.List()
	require.NoError(t, err)
	assert.Equal(t, "Updated event", list[0].Title)

	// Delete
	err = store.Delete(event.ID)
	require.NoError(t, err)

	list, err = store.List()
	require.NoError(t, err)
	assert.Empty(t, list)
}

// бизнес-ошибки
func TestInmemoryStorage_BusinessErrors(t *testing.T) {
	t.Run("Ошибка: Событие не найдено при обновлении", func(t *testing.T) {
		store := New()
		event := &storage.Event{ID: 93, Title: "NotExistingEvent"}
		err := store.Update(event)
		assert.ErrorIs(t, err, storage.ErrEventNotFound)
	})

	t.Run("Ошибка: Событие не найдено при удалении", func(t *testing.T) {
		store := New()
		err := store.Delete(14)
		assert.ErrorIs(t, err, storage.ErrEventNotFound)
	})

	t.Run("Ошибка: Пересечение дат при создании (ErrDateBusy)", func(t *testing.T) {
		store := New()
		now := time.Now()

		event1 := &storage.Event{
			StartTime: now,
			EndTime:   now.Add(1 * time.Hour),
		}
		err := store.Add(event1)
		require.NoError(t, err)

		// Попытка добавить событие, пересекающееся по времени
		event2 := &storage.Event{
			StartTime: now.Add(30 * time.Minute),
			EndTime:   now.Add(2 * time.Hour),
		}
		err = store.Add(event2)
		assert.ErrorIs(t, err, storage.ErrDateBusy)
	})

	t.Run("Ошибка: Пересечение дат при обновлении другого события", func(t *testing.T) {
		store := New()
		now := time.Now()

		event1 := &storage.Event{StartTime: now, EndTime: now.Add(1 * time.Hour)}
		_ = store.Add(event1) // ID: 1

		event2 := &storage.Event{StartTime: now.Add(2 * time.Hour), EndTime: now.Add(3 * time.Hour)}
		_ = store.Add(event2) // ID: 2

		// Пытаемся подвинуть второе событие так, чтобы оно наложилось на первое
		event2.StartTime = now.Add(30 * time.Minute)
		err := store.Update(event2)
		assert.ErrorIs(t, err, storage.ErrDateBusy)
	})
}

// 3. Тест на конкурентную безопасность
func TestInmemoryStorage_Concurrency(t *testing.T) {
	store := New()
	var wg sync.WaitGroup
	workers := 50

	// Запуск параллельных горутин на добавление непересекающихся событий
	// Каждому воркеру даем свое уникальное изолированное окно времени, чтобы не триггерить ErrDateBusy
	baseTime := time.Now()

	for i := 0; i < workers; i++ {
		wg.Add(2)

		// Горутина на запись (Add)
		go func(workerIdx int) {
			defer wg.Done()

			// уникальный временной интервал для каждого воркера
			start := baseTime.Add(time.Duration(workerIdx*2) * time.Hour)
			end := start.Add(1 * time.Hour)

			event := &storage.Event{
				Title:     "Параллельное событие",
				StartTime: start,
				EndTime:   end,
			}
			_ = store.Add(event)
		}(i)

		// Горутина на чтение (List)
		go func() {
			defer wg.Done()
			_, _ = store.List()
		}()
	}

	wg.Wait()

	// Проверяем, что все события успешно записались без паник
	list, err := store.List()
	require.NoError(t, err)
	assert.Len(t, list, workers)
}
