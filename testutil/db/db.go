package testutil

import (
	"context"
	"database/sql"
	"math/rand/v2"
	"testing"
	"time"

	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
)

var mysqlContainer *mysql.MySQLContainer

func init() {
	ctx := context.Background()

	container, err := mysql.Run(ctx, "mysql:8", mysql.WithUsername("root"), mysql.WithPassword("password"))
	if err != nil {
		panic(err)
	}

	conn := container.MustConnectionString(ctx)
	db, err := sql.Open("mysql", conn)
	if err != nil {
		panic(err)
	}
	retry(10, func() error { return db.Ping() })
	db.Close()

	mysqlContainer = container
}

func SetupMysqlDB(t *testing.T, driver string) *sql.DB {
	t.Helper()

	ctx := context.Background()
	connection := mysqlContainer.MustConnectionString(ctx, "parseTime=true")
	db, err := sql.Open(driver, connection)
	if err != nil {
		t.Fatal(err)
	}

	dbName := randomDBName()
	if err := createDatabase(db, dbName); err != nil {
		t.Fatal(err)
	}

	return db
}

type Option func(cfg *mysqldriver.Config) *mysqldriver.Config

func SetUpIsucDB(t *testing.T, setup func(db *sql.DB) error, opts ...Option) *sql.DB {
	t.Helper()

	connection := mysqlContainer.MustConnectionString(context.Background(), "parseTime=true", "multiStatements=true")
	db, err := sql.Open("mysql", connection)
	if err != nil {
		t.Fatal(err)
	}

	dbName := randomDBName()
	if err := createDatabase(db, dbName); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("USE " + dbName); err != nil {
		t.Fatal(err)
	}

	if err := setup(db); err != nil {
		t.Fatal(err)
	}
	db.Close()

	cfg, _ := mysqldriver.ParseDSN(connection)
	cfg.DBName = dbName
	cfg.ParseTime = true
	for _, opt := range opts {
		opt(cfg)
	}

	db, err = sql.Open("mysql+cache", cfg.FormatDSN())
	if err != nil {
		t.Fatal(err)
	}

	return db
}

func WithInterpolateParams() func(cfg *mysqldriver.Config) *mysqldriver.Config {
	return func(cfg *mysqldriver.Config) *mysqldriver.Config {
		cfg.InterpolateParams = true
		return cfg
	}
}

func createDatabase(db *sql.DB, name string) error {
	_, err := db.Exec("CREATE DATABASE " + name)
	if err != nil {
		return err
	}
	_, err = db.Exec("USE " + name)
	return err
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
