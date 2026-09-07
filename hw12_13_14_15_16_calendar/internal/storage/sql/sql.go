package sqlstorage

import (
	"context"
	"fmt"

	"github.com/TheTeaToper/OTUS/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/TheTeaToper/OTUS/hw12_13_14_15_16_calendar/internal/storage"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

// var lgr *logger.Logger

type SqlStorage struct {
	db     *sqlx.DB
	dsn    string
	logger *logger.Logger
}

func New(dsn string, logger *logger.Logger) *SqlStorage {
	logger.Debug("Sql storage. Initialization...")
	return &SqlStorage{dsn: dsn, logger: logger}
}

func (ss *SqlStorage) Connect(ctx context.Context) error {
	ss.logger.Debug("Sql storage. Connecting...")
	db, err := sqlx.Connect("pgx", ss.dsn)
	if err != nil {
		return fmt.Errorf("database connection error: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return fmt.Errorf("ping db error: %w", err)
	}
	ss.db = db
	ss.logger.Debug("Sql storage. Connecting succeeded")
	return nil
}

func (ss *SqlStorage) Close(ctx context.Context) error {
	if ss.db == nil {
		return nil
	}
	ss.logger.Debug("Sql storage. Closing connection...")
	if err := ss.db.Close(); err != nil {
		return fmt.Errorf("closing db error: %w", err)
	}
	ss.logger.Debug("Sql storage. Connection closed")
	return nil
}

func (ss *SqlStorage) Add(e *storage.Event) error {
	var exists bool
	ss.logger.Debug(fmt.Sprintf("Sql storage. Adding event [%#v]...", e))

	checkExistanceQuery := `SELECT EXISTS(SELECT 1 FROM events WHERE start_time < $2 AND end_time > $1)`
	err := ss.db.Get(&exists, checkExistanceQuery, e.StartTime, e.EndTime)
	if err != nil {
		return err
	}
	if exists {
		ss.logger.Debug(fmt.Sprintf("Sql storage. Event [%#v] time slot is busy", e))
		return storage.ErrDateBusy
	}

	insertQuery := `INSERT INTO events (title, start_time, end_time, user_id) VALUES ($1, $2, $3, $4) RETURNING id`
	if err := ss.db.QueryRow(insertQuery, e.Title, e.StartTime, e.EndTime, e.User).Scan(&e.ID); err != nil {
		return fmt.Errorf("inserting event error: %w", err)
	}
	ss.logger.Debug(fmt.Sprintf("Sql storage. Event [%#v] added", e))
	return nil
}

func (ss *SqlStorage) Update(e *storage.Event) error {
	var exists bool
	ss.logger.Debug(fmt.Sprintf("Sql storage. Updating event [%#v]...", e))
	checkExistanceQuery := `SELECT EXISTS(SELECT 1 FROM events WHERE id != $1 AND start_time < $3 AND end_time > $2)`
	err := ss.db.Get(&exists, checkExistanceQuery, e.ID, e.StartTime, e.EndTime)
	if err != nil {
		return err
	}
	if exists {
		ss.logger.Debug(fmt.Sprintf("Sql storage. Event [%#v] time slot is busy", e))
		return storage.ErrDateBusy
	}

	updateQuery := `UPDATE events SET title=$1, start_time=$2, end_time=$3, user_id=$4 WHERE id=$5`
	res, err := ss.db.Exec(updateQuery, e.Title, e.StartTime, e.EndTime, e.User, e.ID)
	if err != nil {
		return fmt.Errorf("updating event error: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return storage.ErrEventNotFound
	}
	ss.logger.Debug(fmt.Sprintf("Sql storage. Event [%#v] updated", e))
	return nil
}

func (ss *SqlStorage) Delete(id int64) error {
	ss.logger.Debug(fmt.Sprintf("Sql storage. Deleting event [ID = %v]...", id))
	deleteQuery := `DELETE FROM events WHERE id=$1`
	res, err := ss.db.Exec(deleteQuery, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return storage.ErrEventNotFound
	}
	ss.logger.Debug(fmt.Sprintf("Sql storage. Event [ID = %v] deleted", id))
	return nil
}

func (ss *SqlStorage) List() ([]storage.Event, error) {
	ss.logger.Debug("Sql storage. Getting events list...")
	selectQuery := `SELECT id, title, start_time, end_time, user_id FROM events`
	var events []storage.Event
	err := ss.db.Select(&events, selectQuery)
	if err != nil {
		return nil, fmt.Errorf("selecting events error: %w", err)
	}
	ss.logger.Debug("Sql storage. Select query completed")
	return events, nil
}
