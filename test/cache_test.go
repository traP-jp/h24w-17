package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traP-jp/isuc/normalizer"
	"github.com/traP-jp/isuc/test/cache"
)

func TestSimpleQuery(t *testing.T) {
	cache.PurgeAllCaches()
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
