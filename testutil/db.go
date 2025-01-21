package testutil

import (
	"context"
	"database/sql"
	"math/rand/v2"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
)

var mysqlContainer *mysql.MySQLContainer

func InitialSetupDB(m *testing.M) {
	ctx := context.Background()

	container, err := mysql.Run(ctx, "mysql:8", mysql.WithUsername("root"), mysql.WithPassword("password"))
	if err != nil {
		panic(err)
	}
	defer container.Terminate(ctx)

	conn := container.MustConnectionString(ctx)
	db, err := sql.Open("mysql", conn)
	if err != nil {
		panic(err)
	}
	retry(10, func() error { return db.Ping() })
	db.Close()

	mysqlContainer = container
	m.Run()
}

func SetupMysqlDB(t *testing.T, driver string) *sql.DB {
	t.Helper()

	ctx := context.Background()

	dbName := randomDBName()
	if err := createDatabase(dbName, driver); err != nil {
		t.Fatal(err)
	}
	connection := mysqlContainer.MustConnectionString(ctx, "parseTime=true")
	db, err := sql.Open(driver, connection)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("USE " + dbName); err != nil {
		t.Fatal(err)
	}

	return db
}

func createDatabase(name string, driver string) error {
	connection := mysqlContainer.MustConnectionString(context.Background(), "multiStatements=true")
	db, err := sql.Open(driver, connection)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("CREATE DATABASE " + name)
	if err != nil {
		return err
	}
	if _, err := db.Exec("USE " + name); err != nil {
		return err
	}

	return nil
}

func randomDBName() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var buf [10]byte
	for i := range buf {
		buf[i] = charset[rand.IntN(len(charset))]
	}
	return "test_" + string(buf[:])
}

func retry(n int, f func() error) {
	var err error
	for i := 0; i < n; i++ {
		err = f()
		if err == nil {
			return
		}
		time.Sleep(1 * time.Second)
	}
	panic(err)
}
