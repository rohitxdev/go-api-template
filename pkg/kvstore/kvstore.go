package kvstore

import (
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/rohitxdev/go-api-starter/internal/common"
)

var (
	ErrKeyNotFound = errors.New("key not found")
	ErrKeyExpired  = errors.New("key expired")
)

type KVStore struct {
	db         *sql.DB
	getStmt    *sql.Stmt
	setStmt    *sql.Stmt
	deleteStmt *sql.Stmt
}

// [db] must be an sqlite3 database
func New(db *sql.DB, purgeFreq time.Duration) (*KVStore, error) {
	var err error
	if _, err = db.Exec("CREATE TABLE IF NOT EXISTS kv_store(key TEXT PRIMARY KEY, value TEXT NOT NULL, expires_at TIMESTAMP);"); err != nil {
		return nil, err
	}

	getStmt, err := db.Prepare("SELECT value, expires_at FROM kv_store WHERE key = $1 AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP);")
	if err != nil {
		return nil, err
	}

	setStmt, err := db.Prepare("INSERT INTO kv_store(key, value, expires_at) VALUES($1, $2, datetime('now', $3)) ON CONFLICT(key) DO UPDATE SET value = $2, expires_at = datetime('now', $3);")
	if err != nil {
		return nil, err
	}

	deleteStmt, err := db.Prepare("DELETE FROM kv_store WHERE key = $1")
	if err != nil {
		return nil, err
	}

	go func() {
		t := time.NewTicker(purgeFreq)
		for {
			<-t.C
			if _, err := db.Exec("DELETE FROM kv_store WHERE expires_at IS NOT NULL AND expires_at < CURRENT_TIMESTAMP;"); err != nil {
				slog.Error("clean up kv store", slog.Any("error", err))
			} else {
				slog.Debug("cleaned up kv store")
			}
		}
	}()

	return &KVStore{
		db:         db,
		getStmt:    getStmt,
		setStmt:    setStmt,
		deleteStmt: deleteStmt,
	}, nil
}

func (kv *KVStore) Close() error {
	var errList []error

	for _, stmt := range []common.Closer{kv.getStmt, kv.setStmt, kv.deleteStmt, kv.db} {
		if err := stmt.Close(); err != nil {
			errList = append(errList, err)
		}
	}

	return errors.Join(errList...)
}

func (kv *KVStore) Get(key string) (string, error) {
	var value string
	var expiresAt time.Time

	err := kv.getStmt.QueryRow(key).Scan(&value, &expiresAt)

	switch {
	case err == sql.ErrNoRows:
		return "", ErrKeyNotFound
	case err != nil:
		return "", err
	case expiresAt.Before(time.Now()):
		return "", ErrKeyExpired
	}

	return value, nil
}

type setOpts struct {
	expiresIn time.Duration
}

func WithExpiry(expiresIn time.Duration) func(*setOpts) {
	return func(so *setOpts) {
		so.expiresIn = expiresIn
	}
}

func (kv *KVStore) Set(key string, value string, optFuncs ...func(*setOpts)) error {
	opts := setOpts{}
	for _, optFunc := range optFuncs {
		optFunc(&opts)
	}

	var expiresAt *time.Time

	if opts.expiresIn > 0 {
		t := time.Now().Add(opts.expiresIn)
		expiresAt = &t
	}
	_, err := kv.setStmt.Exec(key, value, expiresAt)
	return err
}

func (kv *KVStore) Delete(key string) error {
	_, err := kv.deleteStmt.Exec(key)
	return err
}
