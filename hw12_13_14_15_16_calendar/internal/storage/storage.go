package storage

import (
	"errors"
)

var (
	ErrEventNotFound = errors.New("event not found")
	ErrDateBusy      = errors.New("this time slot is already busy by another event")
)

type Storage interface {
	Add(e *Event) error
	Update(e *Event) error
	Delete(id int64) error
	List() ([]Event, error)
}
