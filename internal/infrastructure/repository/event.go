package repository

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	"github.com/kakudo415/warikan-bot/internal/domain/entity"
)

type EventRepository struct {
	db *sql.DB
}

func NewEventRepository(filename string) (*EventRepository, error) {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS events (
			id TEXT PRIMARY KEY
		);
	`)
	if err != nil {
		return nil, err
	}

	return &EventRepository{
		db: db,
	}, nil
}

func (r *EventRepository) CreateIfNotExists(event *entity.Event) error {
	_, err := r.db.Exec("INSERT INTO events (id) VALUES (?) ON CONFLICT IGNORE",
		event.ID.String(),
	)
	return err
}
