package rockscache

import (
	"context"
	"encoding/json"
	"time"
)

type FetchFn[T any] func() (T, error)
type WhenFn[T any] func(T) bool //when to cache

var allTrue = func(_ string) bool { return true }

func Fetch2[T any](ctx context.Context, rc *Client, key string, expire time.Duration, fn FetchFn[T], when WhenFn[T]) (T, error) {
	var fn2 FetchFn[string] = func() (string, error) {
		v, err := fn()
		if err != nil {
			return "", err
		}
		bs, err := json.Marshal(v)
		return string(bs), err
	}

	var when2 WhenFn[string] = func(raw string) bool {
		var v T
		_ = json.Unmarshal([]byte(raw), &v)
		return when(v)
	}

	var res T
	raw, err := rc.Fetch2(ctx, key, expire, fn2, when2)
	if err != nil {
		return res, err
	}
	err = json.Unmarshal([]byte(raw), &res)
	return res, err
}
