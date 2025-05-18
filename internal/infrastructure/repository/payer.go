package repository

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	"github.com/kakudo415/warikan-bot/internal/domain/entity"
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
			id TEXT PRIMARY KEY
		);
	`)
	if err != nil {
		return nil, err
	}

	return &PayerRepository{
		db: db,
	}, nil
}

func (r *PayerRepository) CreatePayer(p *entity.Payer) error {
	_, err := r.db.Exec("INSERT INTO payers (id) VALUES (?)",
		p.ID.String(),
	)
	return err
}
