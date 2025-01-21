package template

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"testing"

	"github.com/traP-jp/h24w-17/testutil"
)

func TestMain(m *testing.M) {
	testutil.InitialSetupDB(m)
}

func TestCacheRows(t *testing.T) {
	db := testutil.SetupMysqlDB(t)
	defer db.Close()

	// setup
	_, err := db.Exec(`INSERT INTO users (id, name) VALUES (1, 'Alice')`)
	if err != nil {
		t.Fatal(err)
	}

	conn, err := db.Conn(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	var cacheRows driver.Rows
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
		cacheRows = NewCachedRows(rows)
		defer cacheRows.Close()

		dest := make([]driver.Value, 2)
		if err := cacheRows.Next(dest); err != nil {
			return err
		}

		id, ok := dest[0].(int64)
		if !ok {
			return fmt.Errorf("dest[0] is not int, got=%T", dest[0])
		}
		if id != 1 {
			return errors.New("id != 1")
		}

		name, ok := dest[1].([]byte)
		if !ok {
			return fmt.Errorf("dest[1] is not string, got=%T", dest[1])
		}
		if string(name) != "Alice" {
			return errors.New("name != Alice")
		}

		return nil
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
