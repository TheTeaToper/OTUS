package storage

import (
	"errors"
)

var (
	ErrEventNotFound = errors.New("event not found")
	ErrDateBusy      = errors.New("this time slot is already busy by another event")
)

type Storage interface {
	CreateEvent(e *Event) error
	UpdateEvent(e *Event) error
	DeleteEvent(id int64) error
	ListEvents() ([]Event, error)
}
