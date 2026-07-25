package memory

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	db *sql.DB
}

type Fact struct {
	Subject   string
	Predicate string
	Object    string
	CreatedAt time.Time
}

func New(path string) (*Store, error) {
	if path == "" {
		path = ":memory:"
	} else {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return nil, fmt.Errorf("create memory dir: %w", err)
		}
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("open memory db: %w", err)
	}

	if err := initSchema(db); err != nil {
		db.Close()
		return nil, err
	}

	return &Store{db: db}, nil
}

func initSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS facts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		subject TEXT NOT NULL,
		predicate TEXT NOT NULL,
		object TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		closed_at DATETIME
	);
	CREATE INDEX IF NOT EXISTS idx_facts_subject ON facts(subject);
	CREATE INDEX IF NOT EXISTS idx_facts_predicate ON facts(predicate);
	`
	_, err := db.Exec(schema)
	return err
}

func (s *Store) Remember(subject, predicate, object string) error {
	_, err := s.db.Exec(
		"INSERT INTO facts (subject, predicate, object) VALUES (?, ?, ?)",
		subject, predicate, object,
	)
	return err
}

func (s *Store) Recall(query string, limit int) ([]Fact, error) {
	if limit <= 0 {
		limit = 10
	}

	rows, err := s.db.Query(
		`SELECT subject, predicate, object, created_at FROM facts
		WHERE closed_at IS NULL AND (
			subject LIKE ? OR predicate LIKE ? OR object LIKE ?
		) ORDER BY created_at DESC LIMIT ?`,
		"%"+query+"%", "%"+query+"%", "%"+query+"%", limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var facts []Fact
	for rows.Next() {
		var f Fact
		if err := rows.Scan(&f.Subject, &f.Predicate, &f.Object, &f.CreatedAt); err != nil {
			return nil, err
		}
		facts = append(facts, f)
	}
	return facts, nil
}

func (s *Store) Forget(subject, predicate string) error {
	_, err := s.db.Exec(
		`UPDATE facts SET closed_at = CURRENT_TIMESTAMP
		WHERE subject = ? AND predicate = ? AND closed_at IS NULL`,
		subject, predicate,
	)
	return err
}

func (s *Store) All() ([]Fact, error) {
	rows, err := s.db.Query(
		"SELECT subject, predicate, object, created_at FROM facts WHERE closed_at IS NULL ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var facts []Fact
	for rows.Next() {
		var f Fact
		if err := rows.Scan(&f.Subject, &f.Predicate, &f.Object, &f.CreatedAt); err != nil {
			return nil, err
		}
		facts = append(facts, f)
	}
	return facts, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}
