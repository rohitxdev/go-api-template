package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

var (
	ErrKeyNotFound = errors.New("key not found")
	ErrKeyExpired  = errors.New("key expired")
)

type KV struct {
	db                *sql.DB
	getStmt           *sql.Stmt
	setStmt           *sql.Stmt
	setWithExpiryStmt *sql.Stmt
	deleteStmt        *sql.Stmt
	name              string
	cleanUpFreq       time.Duration
}

type KVOpts struct {
	CleanUpFreq time.Duration
	InMemory    bool
}

func NewKV(name string, opts KVOpts) (*KV, error) {
	if opts.InMemory {
		name = inMemoryDbName
	}
	db, err := NewDB(name)
	if err != nil {
		return nil, err
	}

	if _, err = db.Exec("CREATE TABLE IF NOT EXISTS kv(key TEXT PRIMARY KEY, value TEXT NOT NULL, expires_at TIMESTAMP);"); err != nil {
		return nil, err
	}

	getStmt, err := db.Prepare("SELECT value, expires_at FROM kv WHERE key = $1 AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP);")
	if err != nil {
		return nil, err
	}

	setStmt, err := db.Prepare("INSERT INTO kv(key, value) VALUES($1, $2) ON CONFLICT(key) DO UPDATE SET value = $2;")
	if err != nil {
		return nil, err
	}

	setWithExpiryStmt, err := db.Prepare("INSERT INTO kv(key, value, expires_at) VALUES($1, $2, datetime('now', $3)) ON CONFLICT(key) DO UPDATE SET value = $2, expires_at = datetime('now', $3);")
	if err != nil {
		return nil, err
	}

	deleteStmt, err := db.Prepare("DELETE FROM kv WHERE key = $1")
	if err != nil {
		return nil, err
	}

	go func() {
		t := time.NewTicker(opts.CleanUpFreq)
		for {
			<-t.C
			if _, err := db.Exec("DELETE FROM kv WHERE expires_at IS NOT NULL AND expires_at < CURRENT_TIMESTAMP;"); err != nil {
				slog.Error("clean up kv store "+name, slog.Any("error", err))
			} else {
				slog.Debug("cleaned up kv store " + name)
			}
		}
	}()

	return &KV{
		db:                db,
		name:              name,
		getStmt:           getStmt,
		setStmt:           setStmt,
		setWithExpiryStmt: setWithExpiryStmt,
		deleteStmt:        deleteStmt,
		cleanUpFreq:       opts.CleanUpFreq,
	}, nil
}

func (kv *KV) Close() error {
	kv.getStmt.Close()
	kv.setStmt.Close()
	kv.setWithExpiryStmt.Close()
	kv.deleteStmt.Close()
	return kv.db.Close()
}

func (kv *KV) Get(key string) (string, error) {
	var value string
	var expiresAt time.Time
	err := kv.getStmt.QueryRow(key).Scan(&value, &expiresAt)
	if err == sql.ErrNoRows {
		return "", ErrKeyNotFound
	}
	if err != nil {
		return "", err
	}
	if expiresAt.Before(time.Now()) {
		return "", ErrKeyExpired
	}
	return value, nil
}

func (kv *KV) Set(key string, value string) error {
	_, err := kv.setStmt.Exec(key, value)
	return err
}

func (kv *KV) Delete(key string) error {
	_, err := kv.deleteStmt.Exec(key)
	return err
}

func (kv *KV) SetWithExpiry(key string, value string, expiresIn time.Duration) error {
	durationStr := fmt.Sprintf("+%d seconds", int(expiresIn.Seconds()))
	_, err := kv.setWithExpiryStmt.Exec(key, value, durationStr, value, durationStr)
	return err
}
