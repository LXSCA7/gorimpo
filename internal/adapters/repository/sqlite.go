package repository

import (
	"database/sql"
	"fmt"

	"github.com/LXSCA7/gorimpo/internal/core/domain"
	"github.com/LXSCA7/gorimpo/internal/core/ports"
	_ "modernc.org/sqlite"
)

type SQLiteRepository struct {
	db *sql.DB
}

var _ ports.OfferRepository = (*SQLiteRepository)(nil)
var _ ports.SystemRepository = (*SQLiteRepository)(nil)

func NewSQLite(dbPath string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	queryOffers := `
	CREATE TABLE IF NOT EXISTS offers (
		link TEXT PRIMARY KEY,
		title TEXT,
		price REAL,
		source TEXT,
		image_url TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := db.Exec(queryOffers); err != nil {
		return nil, fmt.Errorf("error creating offers table: %v", err)
	}

	createDiscardedTable := `
	CREATE TABLE IF NOT EXISTS discarded_offers (
		link TEXT PRIMARY KEY,
		title TEXT,
		price REAL,
		source TEXT,
		image_url TEXT,
		reason TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := db.Exec(createDiscardedTable); err != nil {
		return nil, fmt.Errorf("error creating discarded_offers table: %v", err)
	}

	queryRoutes := `CREATE TABLE IF NOT EXISTS routes (category TEXT PRIMARY KEY, dest_id TEXT);`
	if _, err := db.Exec(queryRoutes); err != nil {
		return nil, fmt.Errorf("error creating routes table: %v", err)
	}

	querySys := `CREATE TABLE IF NOT EXISTS system_info (key TEXT PRIMARY KEY, value TEXT);`
	if _, err := db.Exec(querySys); err != nil {
		return nil, fmt.Errorf("error creating system_info table: %v", err)
	}

	return &SQLiteRepository{db: db}, nil
}

// offers repo:
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

func (r *SQLiteRepository) SaveDiscarded(offer domain.Offer, reason string) (bool, error) {
	query := `INSERT INTO discarded_offers (link, title, price, reason) 
	VALUES (?, ?, ?, ?) 
	ON CONFLICT(link) DO NOTHING`

	result, err := r.db.Exec(query, offer.Link, offer.Title, offer.Price, reason)
	if err != nil {
		return false, err
	}

	// check if row was actually inserted or ignored
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rows > 0, nil
}

// system repo:
func (r *SQLiteRepository) GetRoute(category string) string {
	var dest string
	r.db.QueryRow(`SELECT dest_id FROM routes WHERE category = ?`, category).Scan(&dest)
	return dest
}

func (r *SQLiteRepository) SaveRoute(category, destID string) error {
	query := `INSERT INTO routes (category, dest_id) VALUES (?, ?) ON CONFLICT(category) DO UPDATE SET dest_id = ?`
	_, err := r.db.Exec(query, category, destID, destID)
	return err
}

func (r *SQLiteRepository) GetLastVersion() string {
	var v string
	r.db.QueryRow(`SELECT value FROM system_info WHERE key = 'version'`).Scan(&v)
	return v
}

func (r *SQLiteRepository) SetCurrentVersion(v string) error {
	query := `INSERT INTO system_info (key, value) VALUES ('version', ?) ON CONFLICT(key) DO UPDATE SET value = ?`
	_, err := r.db.Exec(query, v, v)
	return err
}
