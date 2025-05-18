package repository

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	"github.com/kakudo415/warikan-bot/internal/domain/entity"
)

type PaymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(filename string) (*EventRepository, error) {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS payments (
			id TEXT PRIMARY KEY,
			event_id TEXT NOT NULL,
			payer_id TEXT NOT NULL,
			amount INTEGER NOT NULL
		);
	`)
	if err != nil {
		return nil, err
	}

	return &EventRepository{
		db: db,
	}, nil
}

func (r *EventRepository) CreatePayment(p *entity.Payment) error {
	_, err := r.db.Exec("INSERT INTO payments (id, event_id, payer_id, amount) VALUES (?, ?, ?, ?)",
		p.ID.String(),
		p.EventID.String(),
		p.PayerID.String(),
		p.Amount.Uint64(),
	)
	return err
}
