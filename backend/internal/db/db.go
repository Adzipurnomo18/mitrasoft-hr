package db

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"
)

func Open(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	paths := []string{
		"schema.sql",
		filepath.Join("..", "schema.sql"),
	}
	var schema []byte
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err == nil {
			schema = b
			break
		}
	}
	if len(schema) > 0 {
		if _, err := db.Exec(string(schema)); err != nil {
			return nil, err
		}
	}
	return db, nil
}
