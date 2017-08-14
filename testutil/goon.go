package testutil

import (
	"golang.org/x/net/context"

	"github.com/mjibson/goon"

	"google.golang.org/appengine/memcache"
)

// FlushGoonCache は goon のキャッシュをリセットします。
// goon 以外のキャッシュもリセットされることに注意してください。
func FlushGoonCache(ctx context.Context) {
	g := goon.FromContext(ctx)
	g.FlushLocalCache()
	if err := memcache.Flush(ctx); err != nil {
		panic(err)
	}
}
