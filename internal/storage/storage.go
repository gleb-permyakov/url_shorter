package storage

import (
	"database/sql"
	"errors"
	"fmt"

	// "fmt"

	_ "github.com/lib/pq" // init driver
)

var (
	ErrURLNotFound = errors.New("url not found")
	ErrURLExists   = errors.New("url exists or smth")
)

type Storage struct {
	db *sql.DB
}

func New() (*Storage, error) {
	// const op = "storage.postgresql.New" // operation

	connStr := "user=postgres password=root dbname=url_storage sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	// creating db-table
	err = createTable(db)
	if err != nil {
		return &Storage{}, err
	}

	return &Storage{db}, nil
}

func (s *Storage) CloseDb() {
	s.db.Close()
}

func createTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS url (
			id SERIAL PRIMARY KEY,
			alias TEXT NOT NULL UNIQUE,
			url TEXT NOT NULL);
		CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
	`)
	if err != nil {
		return fmt.Errorf("error while creating table: %v", err)
	}
	return nil
}

func (s *Storage) SaveURL(urlToSave string, alias string) (int64, error) {
	const op = "storage.SaveURL"
	// добавляем данные с возвращением индекса
	var id int64
	err := s.db.QueryRow("INSERT INTO url(url, alias) VALUES($1, $2) RETURNING id", urlToSave, alias).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, ErrURLExists)
	}
	// возвращаем индекс
	return id, nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.GetUrl"
	// делаем запрос в бд
	var orig_url string
	err := s.db.QueryRow("SELECT url FROM url WHERE alias = $1", alias).Scan(&orig_url)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrURLNotFound
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}
	// возвращаем оригинальный url
	return orig_url, nil
}

func (s *Storage) DeleteURL(alias string) error {
	const op = "storage.DeleteURL"
	_, err := s.db.Exec("DELETE FROM url WHERE alias = $1", alias)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
