package test

import (
	"database/sql"
	_ "embed"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/traP-jp/isuc/test/cache"
	dbtest "github.com/traP-jp/isuc/testutil/db"
)

//go:embed testdata/0_schema.sql
var tableSchema string

func NewDB(t *testing.T, opts ...dbtest.Option) *sqlx.DB {
	setup := func(db *sql.DB) error {
		_, err := db.Exec(tableSchema)
		if err != nil {
			return err
		}

		for _, user := range InitialData {
			if user.GroupID.Valid {
				_, err := db.Exec(
					"INSERT INTO `users` (`id`, `name`, `age`, `group_id`, `created_at`) VALUES (?, ?, ?, ?, ?)",
					user.ID, user.Name, user.Age, user.GroupID.V, user.CreatedAt,
				)
				if err != nil {
					return err
				}
			} else {
				_, err := db.Exec(
					"INSERT INTO `users` (`id`, `name`, `age`, `created_at`) VALUES (?, ?, ?, ?)",
					user.ID, user.Name, user.Age, user.CreatedAt,
				)
				if err != nil {
					return err
				}
			}
		}

		return nil
	}

	db := dbtest.SetUpIsucDB(t, setup, opts...)

	return sqlx.NewDb(db, "mysql+cache")
}
