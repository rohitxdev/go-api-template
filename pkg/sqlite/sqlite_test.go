package sqlite_test

import (
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/rohitxdev/go-api-starter/pkg/sqlite"
	"github.com/stretchr/testify/assert"
)

func TestSqlite(t *testing.T) {
	dbName := "test_db"
	t.Run("Create DB", func(t *testing.T) {
		db, err := sqlite.NewDB(dbName)
		assert.Nil(t, err)
		defer db.Close()
	})

	t.Cleanup(func() {
		os.RemoveAll(dbName)
	})
}

func TestKV(t *testing.T) {
	var kv *sqlite.KV
	var err error

	kvName := "test_kv"
	t.Run("Create KV store", func(t *testing.T) {
		kv, err = sqlite.NewKV(kvName, sqlite.KVOpts{CleanUpFreq: time.Second * 10})
		assert.Nil(t, err)
	})

	t.Run("Set key", func(t *testing.T) {
		assert.Nil(t, kv.Set("key", "value"))

		value, err := kv.Get("key")
		assert.Equal(t, value, "value")
		assert.Nil(t, err)

		assert.Equal(t, value, "value")
	})

	t.Run("Get key", func(t *testing.T) {
		value, err := kv.Get("key")
		assert.Nil(t, err)
		assert.Equal(t, value, "value")
	})

	t.Run("Delete key", func(t *testing.T) {
		//Confirm key exists before deleting it
		value, err := kv.Get("key")
		assert.NotEqual(t, value, "")
		assert.False(t, errors.Is(err, sql.ErrNoRows))

		assert.Nil(t, kv.Delete("key"))

		value, err = kv.Get("key")
		assert.Equal(t, value, "")
		assert.True(t, errors.Is(err, sql.ErrNoRows))
	})

	t.Cleanup(func() {
		os.RemoveAll(kvName)
	})
}
