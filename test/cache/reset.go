package cache

import (
	"time"

	"github.com/motoki317/sc"
)

func ResetCache() {
	for key := range caches {
		v := caches[key]
		*v.Cache = *sc.NewMust(replaceFn, 10*time.Minute, 10*time.Minute)
		caches[key] = v
	}
}
