package app

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/TheTeaToper/OTUS/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/TheTeaToper/OTUS/hw12_13_14_15_16_calendar/internal/storage"
)

type App struct {
	logger  *logger.Logger
	storage storage.Storage
}

type Storage interface{}

func New(logger *logger.Logger, storage storage.Storage) *App {
	return &App{
		logger:  logger,
		storage: storage,
	}
}

func (a *App) CreateEvent(ctx context.Context, event *storage.Event) error {
	a.logger.Debug("Создание события %s", event.Title)
	return a.storage.CreateEvent(event)
}

func (a *App) UpdateEvent(ctx context.Context, event *storage.Event) error {
	a.logger.Debug("Изменение события %s", event.Title)
	return a.storage.UpdateEvent(event)
}

func (a *App) DeleteEvent(ctx context.Context, id string) error {
	a.logger.Debug("Удаление события с ID %s", id)

	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return fmt.Errorf("невалидный формат ID: %w", err)
	}

	return a.storage.DeleteEvent(idInt)
}

func (a *App) ListEventsForDay(ctx context.Context, day time.Time) ([]storage.Event, error) {
	a.logger.Debug("Получение списка событий на день %s", day.Format("2006-01-02"))

	allEvents, err := a.storage.ListEvents()
	if err != nil {
		return nil, err
	}

	var filteredEvents []storage.Event
	for _, e := range allEvents {
		if e.StartTime.Year() == day.Year() && e.StartTime.YearDay() == day.YearDay() {
			filteredEvents = append(filteredEvents, e)
		}
	}
	return filteredEvents, nil
}

func (a *App) ListEventsForWeek(ctx context.Context, weekFirstDay time.Time) ([]storage.Event, error) {
	a.logger.Debug("Получение списка событий на неделю, начиная с %s", weekFirstDay.Format("2006-01-02"))

	allEvents, err := a.storage.ListEvents()
	if err != nil {
		return nil, err
	}

	weekLastDay := weekFirstDay.AddDate(0, 0, 7)
	var filteredEvents []storage.Event
	for _, e := range allEvents {
		if (e.StartTime.After(weekFirstDay) || e.StartTime.Equal(weekFirstDay)) && e.StartTime.Before(weekLastDay) {
			filteredEvents = append(filteredEvents, e)
		}
	}
	return filteredEvents, nil
}

func (a *App) ListEventsForMonth(ctx context.Context, monthFirstDay time.Time) ([]storage.Event, error) {
	a.logger.Debug("Получение списка событий на месяц, начиная с %s", monthFirstDay.Format("2006-01-02"))

	allEvents, err := a.storage.ListEvents()
	if err != nil {
		return nil, err
	}

	monthLastDay := monthFirstDay.AddDate(0, 1, 0)
	var filteredEvents []storage.Event
	for _, e := range allEvents {
		if (e.StartTime.After(monthFirstDay) || e.StartTime.Equal(monthFirstDay)) && e.StartTime.Before(monthLastDay) {
			filteredEvents = append(filteredEvents, e)
		}
	}
	return filteredEvents, nil
}
