package kv

import (
	"database/sql"
)

type KV struct {
	db      *sql.DB
	getStmt *sql.Stmt
	setStmt *sql.Stmt
	name    string
}

func New(name string) (*KV, error) {
	db, err := newSqliteDb(name + ".db")
	if err != nil {
		return nil, err
	}

	if _, err = db.Exec("CREATE TABLE IF NOT EXISTS kv (key TEXT PRIMARY KEY, value TEXT NOT NULL)"); err != nil {
		return nil, err
	}

	getStmt, err := db.Prepare("SELECT value FROM kv WHERE key=$1")
	if err != nil {
		return nil, err
	}
	setStmt, err := db.Prepare("INSERT INTO kv(key, value) VALUES($1, $2) ON CONFLICT(key) DO UPDATE SET value=$2")
	if err != nil {
		return nil, err
	}

	return &KV{
		db:      db,
		name:    name,
		getStmt: getStmt,
		setStmt: setStmt,
	}, nil
}

func (kv *KV) Get(key string) (string, error) {
	var value string
	err := kv.getStmt.QueryRow(key).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrKeyNotFound
		}
		return "", err
	}
	return value, nil
}

func (kv *KV) Set(key string, value string) error {
	_, err := kv.setStmt.Exec(key, value)
	return err
}
