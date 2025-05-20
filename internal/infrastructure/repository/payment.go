package repository

import (
	"database/sql"
	"errors"

	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"

	"github.com/kakudo415/warikan-bot/internal/domain/entity"
	"github.com/kakudo415/warikan-bot/internal/domain/valueobject"
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

func (r *EventRepository) Create(payment *entity.Payment) error {
	_, err := r.db.Exec("INSERT INTO payments (id, event_id, payer_id, amount) VALUES (?, ?, ?, ?)",
		payment.ID.String(),
		payment.EventID.String(),
		payment.PayerID.String(),
		payment.Amount.Uint64(),
	)
	if sqliteErr := new(sqlite3.Error); errors.As(err, sqliteErr) {
		if sqliteErr.ExtendedCode == sqlite3.ErrConstraintPrimaryKey || sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return valueobject.NewErrorAlreadyExists("payment already exists", err)
		}
	}
	return err
}

func (r *EventRepository) Delete(paymentID valueobject.PaymentID) error {
	_, err := r.db.Exec("DELETE FROM payments WHERE id = ?", paymentID.String())
	return err
}

func (r *EventRepository) FindByEventID(eventID valueobject.EventID) ([]*entity.Payment, error) {
	rows, err := r.db.Query("SELECT id, event_id, payer_id, amount FROM payments WHERE event_id = ?", eventID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*entity.Payment
	for rows.Next() {
		var rawID, rawEventID, rawPayerID string
		var rawAmount int
		var payment entity.Payment
		err := rows.Scan(&rawID, &rawEventID, &rawPayerID, &rawAmount)
		if err != nil {
			return nil, err
		}
		payment.ID, err = valueobject.NewPaymentIDFromString(rawID)
		if err != nil {
			return nil, err
		}
		payment.EventID = valueobject.NewEventID(rawEventID)
		payment.PayerID = valueobject.NewPayerID(rawPayerID)
		payment.Amount, err = valueobject.NewYen(rawAmount)
		if err != nil {
			return nil, err
		}
		payments = append(payments, &payment)
	}

	return payments, nil
}
