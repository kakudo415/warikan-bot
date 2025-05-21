package repository

import (
	"database/sql"
	"errors"

	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"

	"github.com/kakudo415/warikan-bot/internal/domain/entity"
	"github.com/kakudo415/warikan-bot/internal/domain/valueobject"
)

type PayerRepository struct {
	db *sql.DB
}

func NewPayerRepository(filename string) (*PayerRepository, error) {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS payers (
			id TEXT PRIMARY KEY,
			event_id TEXT NOT NULL
		);
	`)
	if err != nil {
		return nil, err
	}

	return &PayerRepository{
		db: db,
	}, nil
}

func (r *PayerRepository) Create(payer *entity.Payer) error {
	_, err := r.db.Exec("INSERT INTO payers (id, event_id) VALUES (?, ?)",
		payer.ID.String(),
		payer.EventID.String(),
	)
	if sqliteErr := new(sqlite3.Error); errors.As(err, sqliteErr) {
		if sqliteErr.ExtendedCode == sqlite3.ErrConstraintPrimaryKey || sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return valueobject.NewErrorAlreadyExists("payer already exists", err)
		}
	}
	return err
}

func (r *PayerRepository) CreateIfNotExists(payer *entity.Payer) error {
	_, err := r.db.Exec("INSERT OR IGNORE INTO payers (id, event_id) VALUES (?, ?)",
		payer.ID.String(),
		payer.EventID.String(),
	)
	return err
}

func (r *PayerRepository) FindByEventID(eventID valueobject.EventID) ([]*entity.Payer, error) {
	rows, err := r.db.Query("SELECT id, event_id FROM payers WHERE event_id = ?", eventID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payers []*entity.Payer
	for rows.Next() {
		var rawID, rawEventID string
		var payer entity.Payer
		if err := rows.Scan(&rawID, &rawEventID); err != nil {
			return nil, err
		}
		payer.ID = valueobject.NewPayerID(rawID)
		payer.EventID = valueobject.NewEventID(rawEventID)
		payers = append(payers, &payer)
	}
	return payers, nil
}
