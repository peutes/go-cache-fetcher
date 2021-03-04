package cachefetcher_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/peutes/redis-fetcher/cachefetcher"
)

const host = "localhost:6379"

var options = &cachefetcher.Options{WithStackTrace: true, DebugPrintMode: true}

func getClient() cachefetcher.Client {
	c := redis.NewUniversalClient(
		&redis.UniversalOptions{Addrs: []string{host}},
	)
	return &cachefetcher.SampleCacheClientImpl{
		Client: c,
		Ctx:    context.Background(),
	}
}

func TestClient(t *testing.T) {
	c := getClient()

	if err := c.Set("key", "value", 0); err != nil {
		t.Error(err)
	}

	var val string
	err := c.Get("key", &val)
	if err != nil {
		t.Error(err)
	}
	if val != "value" {
		t.Errorf("failed: %+v", val)
	}

	err = c.Get("key2", &val)
	if !errors.Is(err, redis.Nil) {
		t.Errorf("failed: %+v, %+v", val, err)
	}
}

func TestSetKey(t *testing.T) {
	f := cachefetcher.NewCacheFetcher(getClient(), options)
	f.SetKey([]string{"prefix", "key"}, false, "hoge", "fuga")
	key := f.Key()

	want := "prefix_key_hoge_fuga"
	if key != want {
		t.Errorf("%+v", key)
	}
}

func TestSetKeyWithHash(t *testing.T) {
	f := cachefetcher.NewCacheFetcher(getClient(), options)
	f.SetKey([]string{"prefix", "key"}, true, "hoge", "fugadddddddd")
	key := f.Key()

	want := "prefix_key_a31d03600d04dd35fc74f8489c9347d154074699ddb37ca893f3a0a9e20ac09d"
	if key != want {
		t.Errorf("%+v", key)
	}
}

func TestFetch(t *testing.T) {
	f := cachefetcher.NewCacheFetcher(getClient(), options)
	f.SetKey([]string{"prefix", "key"}, false, "hoge", "fuga")

	// first fetch read from fetcher.
	var dst string
	want := "piyo"
	dst2, err := f.Fetch(10*time.Second, &dst, func() (string, error) {
		return want, nil
	})
	if err != nil {
		t.Errorf("%+v", err)
	}

	if f.IsCached() || dst != "" {
		t.Errorf("%+v %+v", f.IsCached(), dst)
	}

	if dst2 != want {
		t.Errorf("%+v", dst2)
	}

	// second fetch read from cache.
	dst3, err := f.Fetch(10*time.Second, &dst, func() (string, error) {
		return want, nil
	})
	if err != nil {
		t.Errorf("%+v", err)
	}

	if !f.IsCached() {
		t.Errorf("%+v", f.IsCached())
	}

	if dst != want || dst3 != want {
		t.Errorf("%+v, %+v", dst, dst3)
	}
}

func TestSetVal(t *testing.T) {
	f := cachefetcher.NewCacheFetcher(getClient(), options)
	f.SetKey([]string{"prefix", "key"}, false, "hoge", "fuga")
	if err := f.SetVal("value", 10*time.Second); err != nil {
		t.Errorf("%+v", err)
	}
}

func TestGetString(t *testing.T) {
	f := cachefetcher.NewCacheFetcher(getClient(), options)
	f.SetKey([]string{"prefix", "key"}, true, "hoge", "fuga")
	want := "value"
	if err := f.SetVal(want, 10*time.Second); err != nil {
		t.Errorf("%+v", err)
	}

	dst, err := f.GetString()
	if err != nil {
		t.Errorf("%+v", err)
	}

	if !f.IsCached() {
		t.Errorf("%+v", f.IsCached())
	}

	if dst != want {
		t.Errorf("%+v", dst)
	}
}

func TestGetVal(t *testing.T) {
	f := cachefetcher.NewCacheFetcher(getClient(), options)
	f.SetKey([]string{"prefix", "key"}, true, "hoge", "fuga")
	want := "value"
	if err := f.SetVal(want, 10*time.Second); err != nil {
		t.Errorf("%+v", err)
	}

	var dst string
	dst2, err := f.GetVal(&dst)
	if err != nil {
		t.Errorf("%+v", err)
	}

	if !f.IsCached() {
		t.Errorf("%+v", f.IsCached())
	}

	if dst != want || dst2 != want {
		t.Errorf("%+v, %+v", dst, dst2)
	}
}

func TestDelVal(t *testing.T) {
	f := cachefetcher.NewCacheFetcher(getClient(), options)
	f.SetKey([]string{"prefix", "key"}, false, "hoge", "fuga")
	if err := f.SetVal("value", 10*time.Second); err != nil {
		t.Errorf("%+v", err)
	}

	if err := f.DelVal(); err != nil {
		t.Errorf("%+v", err)
	}

	var dst string
	dst2, err := f.GetVal(&dst)
	if err != nil && !errors.Is(err, redis.Nil) {
		t.Errorf("%+v", err)
	}
	if dst != "" || dst2 != nil {
		t.Errorf("%+v, %+v", dst, dst2)
	}
}