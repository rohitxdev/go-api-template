package sqlite_test

import (
	"database/sql"
	"errors"
	"os"
	"testing"

	"github.com/rohitxdev/go-api-template/pkg/assert"
	"github.com/rohitxdev/go-api-template/pkg/sqlite"
)

func TestSqlite(t *testing.T) {
	t.Run("Create DB", func(t *testing.T) {
		db, err := sqlite.NewDB("test_db")
		assert.Nil(t, err)
		defer db.Close()
	})

	t.Cleanup(func() {
		os.RemoveAll("db")
	})
}

func TestKV(t *testing.T) {
	var kv *sqlite.KV
	var err error
	t.Run("Create KV store", func(t *testing.T) {
		kv, err = sqlite.NewKV("test_kv")
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
		os.RemoveAll("db")
	})
}
