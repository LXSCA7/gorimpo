package repository

import (
	"database/sql"
	"fmt"

	"github.com/LXSCA7/gorimpo/internal/core/domain"
	_ "modernc.org/sqlite"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLite(dbPath string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	query := `
	CREATE TABLE IF NOT EXISTS offers (
		link TEXT PRIMARY KEY,
		title TEXT,
		price REAL,
		source TEXT,
		image_url TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := db.Exec(query); err != nil {
		return nil, fmt.Errorf("erro ao criar tabela offers: %v", err)
	}

	return &SQLiteRepository{db: db}, nil
}

func (r *SQLiteRepository) OfferExists(link string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM offers WHERE link = ?)`
	err := r.db.QueryRow(query, link).Scan(&exists)
	return exists, err
}

func (r *SQLiteRepository) SaveOffer(offer domain.Offer) error {
	query := `
	INSERT INTO offers (link, title, price, source, image_url) 
	VALUES (?, ?, ?, ?, ?)`

	_, err := r.db.Exec(query, offer.Link, offer.Title, offer.Price, offer.Source, offer.ImageURL)
	return err
}
