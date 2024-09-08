package sqlite

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"
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

func NewKV(name string, cleanUpFreq time.Duration) (*KV, error) {
	db, err := NewDB(name)
	if err != nil {
		return nil, err
	}

	if _, err = db.Exec("CREATE TABLE IF NOT EXISTS kv(key TEXT PRIMARY KEY, value TEXT NOT NULL, expires_at TIMESTAMP);"); err != nil {
		return nil, err
	}

	getStmt, err := db.Prepare("SELECT value FROM kv WHERE key = ? AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP);")
	if err != nil {
		return nil, err
	}

	setStmt, err := db.Prepare("INSERT INTO kv(key, value) VALUES(?, ?) ON CONFLICT(key) DO UPDATE SET value = ?;")
	if err != nil {
		return nil, err
	}

	setWithExpiryStmt, err := db.Prepare("INSERT INTO kv(key, value, expires_at) VALUES(?, ?, datetime('now', ?)) ON CONFLICT(key) DO UPDATE SET value = ?, expires_at = datetime('now', ?);")
	if err != nil {
		return nil, err
	}

	deleteStmt, err := db.Prepare("DELETE FROM kv WHERE key = ?")
	if err != nil {
		return nil, err
	}

	go func() {
		t := time.NewTicker(cleanUpFreq)
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
		cleanUpFreq:       cleanUpFreq,
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
	err := kv.getStmt.QueryRow(key).Scan(&value)
	return value, err
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
