package repo

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

var (
	ErrKeyNotFound = sql.ErrNoRows
)

func NewSqlite(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	stmts := []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA synchronous=NORMAL;",
		"PRAGMA locking_mode=NORMAL;",
		"PRAGMA busy_timeout=10000;",
		"PRAGMA cache_size=10000;",
	}

	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return nil, err
		}
	}
	return db, nil
}
