package test

import (
	"database/sql"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/traP-jp/isuc/normalizer"
	"github.com/traP-jp/isuc/test/cache"
)

func TestSimpleQuery(t *testing.T) {
	cache.ResetCache()
	db := NewDB(t)

	var user User
	err := db.Get(&user, "SELECT * FROM `users` WHERE `id` = ?", 1)
	if err != nil {
		t.Fatal(err)
	}

	AssertUser(t, InitialData[0], user)

	stats := cache.ExportCacheStats()[normalizer.NormalizeQuery("SELECT * FROM `users` WHERE `id` = ?")]
	assert.Equal(t, 0, stats.Hits)
	assert.Equal(t, 1, stats.Misses)
}

func TestSimpleQueryCache(t *testing.T) {
	cache.ResetCache()
	db := NewDB(t)

	var user User
	err := db.Get(&user, "SELECT * FROM `users` WHERE `id` = ?", 1)
	if err != nil {
		t.Fatal(err)
	}

	AssertUser(t, InitialData[0], user)

	var user2 User
	err = db.Get(&user2, "SELECT * FROM `users` WHERE `id` = ?", 1)
	if err != nil {
		t.Fatal(err)
	}

	AssertUser(t, InitialData[0], user2)

	stats := cache.ExportCacheStats()[normalizer.NormalizeQuery("SELECT * FROM `users` WHERE `id` = ?")]
	assert.Equal(t, 1, stats.Hits)
	assert.Equal(t, 1, stats.Misses)
}

func TestSelectAfterUpdate(t *testing.T) {
	cache.ResetCache()
	db := NewDB(t)

	var user User
	err := db.Get(&user, "SELECT * FROM `users` WHERE `id` = ?", 1)
	if err != nil {
		t.Fatal(err)
	}

	AssertUser(t, InitialData[0], user)

	_, err = db.Exec("UPDATE `users` SET `name` = ? WHERE `id` = ?", "updated", 1)
	if err != nil {
		t.Fatal(err)
	}
	user.Name = "updated"

	// no cache hit because users with id=1 is updated
	var user2 User
	err = db.Get(&user2, "SELECT * FROM `users` WHERE `id` = ?", 1)
	if err != nil {
		t.Fatal(err)
	}

	AssertUser(t, user, user2)

	stats := cache.ExportCacheStats()[normalizer.NormalizeQuery("SELECT * FROM `users` WHERE `id` = ?")]
	assert.Equal(t, 0, stats.Hits)
	assert.Equal(t, 2, stats.Misses)
}

func TestSelectAfterInsert(t *testing.T) {
	cache.ResetCache()
	db := NewDB(t)

	var user User
	err := db.Get(&user, "SELECT * FROM `users` WHERE `id` = ?", 1)
	if err != nil {
		t.Fatal(err)
	}

	AssertUser(t, InitialData[0], user)

	newUser := User{
		Name:      "new",
		Age:       10,
		CreatedAt: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	_, err = db.Exec("INSERT INTO `users` (`name`, `age`, `created_at`) VALUES (?, ?, ?)", newUser.Name, newUser.Age, newUser.CreatedAt)
	if err != nil {
		t.Fatal(err)
	}

	// cache hit because users with id=1 is not updated
	err = db.Get(&user, "SELECT * FROM `users` WHERE `id` = ?", 1)
	if err != nil {
		t.Fatal(err)
	}

	AssertUser(t, InitialData[0], user)

	stats := cache.ExportCacheStats()[normalizer.NormalizeQuery("SELECT * FROM `users` WHERE `id` = ?")]
	assert.Equal(t, 1, stats.Hits)
	assert.Equal(t, 1, stats.Misses)
}

func TestSelectIn(t *testing.T) {
	cache.ResetCache()
	db := NewDB(t)

	var users []User
	err := db.Select(&users, "SELECT * FROM `users` WHERE `id` IN (?, ?)", 1, 2)
	if err != nil {
		t.Fatal(err)
	}

	AssertUser(t, InitialData[0], users[0])
	AssertUser(t, InitialData[1], users[1])

	// cache hit
	err = db.Select(&users, "SELECT * FROM `users` WHERE `id` IN (?, ?)", 1, 2)
	if err != nil {
		t.Fatal(err)
	}

	AssertUser(t, InitialData[0], users[0])
	AssertUser(t, InitialData[1], users[1])

	// the IN query is separately cached
	stats := cache.ExportCacheStats()[normalizer.NormalizeQuery("SELECT * FROM `users` WHERE `id` = ?")]
	assert.Equal(t, 2, stats.Hits)
	assert.Equal(t, 2, stats.Misses)
}

func TestSelectUsersByGroupID(t *testing.T) {
	cache.ResetCache()
	db := NewDB(t)

	var users []User
	err := db.Select(&users, "SELECT * FROM `users` WHERE `group_id` = ?", 1)
	if err != nil {
		t.Fatal(err)
	}

	group1Users := make([]User, 0)
	for _, user := range InitialData {
		if user.GroupID.Valid && user.GroupID.V == 1 {
			group1Users = append(group1Users, user)
		}
	}
	AssertUsers(t, group1Users, users)

	newUser := User{
		Name:      "new",
		Age:       10,
		GroupID:   sql.Null[int]{Valid: true, V: 2},
		CreatedAt: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	_, err = db.Exec(
		"INSERT INTO `users` (`name`, `age`, `group_id`, `created_at`) VALUES (?, ?, ?, ?)",
		newUser.Name, newUser.Age, newUser.GroupID.V, newUser.CreatedAt,
	)
	if err != nil {
		t.Fatal(err)
	}

	// cache hit because users with group_id=1 is not updated
	err = db.Select(&users, "SELECT * FROM `users` WHERE `group_id` = ?", 1)
	if err != nil {
		t.Fatal(err)
	}

	AssertUsers(t, group1Users, users)

	stats := cache.ExportCacheStats()[normalizer.NormalizeQuery("SELECT * FROM `users` WHERE `group_id` = ?")]
	assert.Equal(t, 1, stats.Hits)
	assert.Equal(t, 1, stats.Misses)
}

func TestTransaction(t *testing.T) {
	cache.ResetCache()
	db := NewDB(t)

	errCh := make(chan error, 2)
	afterUpdate := make(chan struct{})
	beforeCommit := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(2)

	// transaction
	go func() {
		defer wg.Done()

		tx, err := db.Beginx()
		if err != nil {
			errCh <- err
			return
		}
		defer tx.Rollback()

		_, err = tx.Exec("UPDATE `users` SET `name` = ? WHERE `id` = ?", "updated", 1)
		if err != nil {
			errCh <- err
			return
		}

		close(afterUpdate)

		<-beforeCommit

		err = tx.Commit()
		errCh <- err

		t.Log("transaction committed")
	}()

	// select
	go func() {
		defer wg.Done()

		<-afterUpdate

		var user User
		err := db.Get(&user, "SELECT * FROM `users` WHERE `id` = ?", 1)
		if err != nil {
			errCh <- err
			return
		}

		// user must not be updated yet because the transaction is not committed
		AssertUser(t, InitialData[0], user)

		close(beforeCommit)

		errCh <- nil

		t.Log("select completed")
	}()

	wg.Wait()
	for range len(errCh) {
		if err := <-errCh; err != nil {
			t.Fatal(err)
		}
	}

	// now user must be updated
	var user User
	err := db.Get(&user, "SELECT * FROM `users` WHERE `id` = ?", 1)
	if err != nil {
		t.Fatal(err)
	}
	updated := InitialData[0]
	updated.Name = "updated"
	AssertUser(t, updated, user)
}

func TestSelectTransaction(t *testing.T) {
	cache.ResetCache()
	db := NewDB(t)

	tx := db.MustBegin()
	defer tx.Rollback()

	var user User
	err := tx.Get(&user, "SELECT * FROM `users` WHERE `id` = ?", 1)
	if err != nil {
		t.Fatal(err)
	}

	AssertUser(t, InitialData[0], user)

	// cache hit
	err = tx.Get(&user, "SELECT * FROM `users` WHERE `id` = ?", 1)
	if err != nil {
		t.Fatal(err)
	}

	AssertUser(t, InitialData[0], user)

	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	stats := cache.ExportCacheStats()[normalizer.NormalizeQuery("SELECT * FROM `users` WHERE `id` = ?")]
	assert.Equal(t, 1, stats.Hits)
	assert.Equal(t, 1, stats.Misses)
}

func TestFuzzyRead(t *testing.T) {
	cache.ResetCache()
	db := NewDB(t)

	errCh := make(chan error, 2)
	afterFirstQuery := make(chan struct{})
	afterUpdate := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(2)

	// transaction 1
	go func() {
		defer wg.Done()

		tx := db.MustBegin()
		defer tx.Rollback()

		var user1 User
		err := tx.Get(&user1, "SELECT * FROM `users` WHERE `id` = ?", 1)
		if err != nil {
			errCh <- err
			return
		}
		AssertUser(t, InitialData[0], user1)

		close(afterFirstQuery)

		<-afterUpdate

		var user2 User
		err = tx.Get(&user2, "SELECT * FROM `users` WHERE `id` = ?", 1)
		if err != nil {
			errCh <- err
			return
		}

		AssertUser(t, user1, user2)

		errCh <- tx.Commit()
	}()

	// transaction 2
	go func() {
		defer wg.Done()

		tx := db.MustBegin()
		defer tx.Rollback()

		<-afterFirstQuery

		_, err := tx.Exec("UPDATE `users` SET `name` = ? WHERE `id` = ?", "updated", 1)
		if err != nil {
			errCh <- err
			return
		}

		err = tx.Commit()
		if err != nil {
			errCh <- err
			return
		}

		close(afterUpdate)
	}()

	wg.Wait()
	for range len(errCh) {
		if err := <-errCh; err != nil {
			t.Fatal(err)
		}
	}
}
