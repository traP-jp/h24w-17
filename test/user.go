package test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type User struct {
	ID        int           `db:"id"`
	Name      string        `db:"name"`
	Age       int           `db:"age"`
	GroupID   sql.Null[int] `db:"group_id"`
	CreatedAt time.Time     `db:"created_at"`
}

func AssertUser(t *testing.T, expected, actual User) {
	assert.Equal(t, expected.ID, actual.ID, "ID")
	assert.Equal(t, expected.Name, actual.Name, "Name")
	assert.Equal(t, expected.Age, actual.Age, "Age")
	assert.Equal(t, expected.GroupID, actual.GroupID, "GroupID")
	assert.WithinDuration(t, expected.CreatedAt, actual.CreatedAt, 1*time.Second, "CreatedAt")
}

var InitialData = []User{
	{
		ID:        1,
		Name:      "Alice",
		Age:       20,
		GroupID:   sql.Null[int]{Valid: true, V: 1},
		CreatedAt: time.Now(),
	},
	{
		ID:        2,
		Name:      "Bob",
		Age:       21,
		GroupID:   sql.Null[int]{Valid: true, V: 1},
		CreatedAt: time.Now(),
	},
	{
		ID:        3,
		Name:      "Charlie",
		Age:       22,
		GroupID:   sql.Null[int]{},
		CreatedAt: time.Now(),
	},
	{
		ID:        4,
		Name:      "David",
		Age:       23,
		GroupID:   sql.Null[int]{Valid: true, V: 2},
		CreatedAt: time.Now(),
	},
}
