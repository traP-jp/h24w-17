package template

import (
	"context"
	"database/sql/driver"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traP-jp/isuc/testutil"
)

func TestMain(m *testing.M) {
	testutil.InitialSetupDB(m)
}

func TestCacheRows(t *testing.T) {
	db := testutil.SetupMysqlDB(t, "mysql")
	defer db.Close()

	schema, err := os.ReadFile("testdata/schema.sql")
	assert.NoError(t, err)
	_, err = db.Exec(string(schema))
	assert.NoError(t, err)

	// setup
	_, err = db.Exec(`INSERT INTO users (id, name) VALUES (1, 'Alice')`)
	assert.NoError(t, err)

	conn, err := db.Conn(context.Background())
	assert.NoError(t, err)
	defer conn.Close()

	var cacheRows *cacheRows
	err = conn.Raw(func(driverConn any) error {
		conn := driverConn.(driver.Conn)
		stmt, err := conn.Prepare("SELECT id, name FROM users")
		if err != nil {
			return err
		}
		defer stmt.Close()

		rows, err := stmt.Query(nil)
		if err != nil {
			return err
		}
		cacheRows, err = newCacheRows(rows)
		return err
	})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// check cache
	checkCache := func() {
		defer cacheRows.Close()

		dest := make([]driver.Value, 2)
		if err := cacheRows.Next(dest); err != nil {
			t.Error(err)
			t.FailNow()
		}

		id, ok := dest[0].(int64)
		if !ok {
			t.Errorf("dest[0] is not int, got=%T", dest[0])
		}
		if id != 1 {
			t.Error("id != 1")
		}
		name, ok := dest[1].([]byte)
		if !ok {
			t.Errorf("dest[1] is not string, got=%T", dest[1])
		}
		if string(name) != "Alice" {
			t.Error("name != Alice")
		}
	}

	checkCache()
	checkCache()
}
