package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"

	_ "modernc.org/sqlite"
)

var (
	ErrIsNotDir  = errors.New("is not a directory")
	ErrCreateDir = errors.New("could not create directory")
	ErrStatDir   = errors.New("could not stat directory")
)

func createDirIfNotExists(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			if err = os.Mkdir(path, 0755); err != nil {
				return errors.Join(ErrCreateDir, err)
			}
			slog.Debug(fmt.Sprintf("created directory %s ✔︎", path))
		} else {
			// Other error
			return errors.Join(ErrStatDir, err)
		}
	} else if !info.IsDir() {
		return errors.Join(ErrIsNotDir, err)
	}

	return nil
}

func NewDB(name string) (*sql.DB, error) {
	dirName := "db"
	if err := createDirIfNotExists(dirName); err != nil && err != ErrIsNotDir {
		return nil, err
	}
	db, err := sql.Open("sqlite", fmt.Sprintf("%s/%s.db", dirName, name))
	if err != nil {
		return nil, err
	}

	stmts := [...]string{
		"PRAGMA journal_mode = WAL;",
		"PRAGMA synchronous = NORMAL;",
		"PRAGMA locking_mode = NORMAL;",
		"PRAGMA busy_timeout = 10000;",
		"PRAGMA cache_size = 10000;",
		"PRAGMA foreign_keys = ON;",
	}

	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return nil, err
		}
	}
	return db, nil
}
