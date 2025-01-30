package test

import (
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
