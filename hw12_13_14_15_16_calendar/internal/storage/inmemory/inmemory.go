package inmemorystorage

import (
	"sync"

	"github.com/TheTeaToper/OTUS/hw12_13_14_15_16_calendar/internal/storage"
)

type InmemoryStorage struct {
	mutex  sync.RWMutex
	events map[int64]storage.Event
	nextID int64
}

func New() *InmemoryStorage {
	return &InmemoryStorage{
		events: make(map[int64]storage.Event),
		nextID: 1,
	}
}

func (ms *InmemoryStorage) Add(e *storage.Event) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	for _, event := range ms.events {
		if e.StartTime.Before(event.EndTime) && e.EndTime.After(event.StartTime) {
			return storage.ErrDateBusy
		}
	}

	e.ID = ms.nextID
	ms.nextID++
	ms.events[e.ID] = *e
	return nil
}

func (ms *InmemoryStorage) Update(e *storage.Event) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	if _, exists := ms.events[e.ID]; !exists {
		return storage.ErrEventNotFound
	}

	for _, event := range ms.events {
		if event.ID == e.ID {
			continue
		}
		if e.StartTime.Before(event.EndTime) && e.EndTime.After(event.StartTime) {
			return storage.ErrDateBusy
		}
	}

	ms.events[e.ID] = *e
	return nil
}

func (ms *InmemoryStorage) Delete(id int64) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	if _, exists := ms.events[id]; !exists {
		return storage.ErrEventNotFound
	}

	delete(ms.events, id)
	return nil
}

func (ms *InmemoryStorage) List() ([]storage.Event, error) {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	list := make([]storage.Event, 0, len(ms.events))
	for _, event := range ms.events {
		list = append(list, event)
	}
	return list, nil
}
