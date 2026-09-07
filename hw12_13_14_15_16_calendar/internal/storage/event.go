package storage

import "time"

type Event struct {
	ID        int64     `db:"id" json:"id"`
	Title     string    `db:"title" json:"title"`
	StartTime time.Time `db:"start_time" json:"start_time"`
	EndTime   time.Time `db:"end_time" json:"end_time"`
	User      string    `db:"user_id" json:"user_id"`
}
