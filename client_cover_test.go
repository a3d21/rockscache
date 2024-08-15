package rockscache

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBadOptions(t *testing.T) {
	assert.Panics(t, func() {
		NewClient(nil, Options{})
	})
}

func TestDisable(t *testing.T) {
	rc := NewClient(nil, NewDefaultOptions())
	rc.Options.DisableCacheDelete = true
	rc.Options.DisableCacheRead = true
	fn := func() (string, error) { return "", nil }
	_, err := rc.Fetch2(context.Background(), "key", 60, fn, allTrue)
	assert.Nil(t, err)
	err = rc.TagAsDeleted2(context.Background(), "key")
	assert.Nil(t, err)
}

func TestEmptyExpire(t *testing.T) {
	testEmptyExpire(t, 0)
	testEmptyExpire(t, 10*time.Second)
}

func testEmptyExpire(t *testing.T, expire time.Duration) {
	clearCache()
	rc := NewClient(rdb, NewDefaultOptions())
	rc.Options.EmptyExpire = expire
	fn := func() (string, error) { return "", nil }
	fetchError := errors.New("fetch error")
	errFn := func() (string, error) {
		return "", fetchError
	}
	_, err := rc.Fetch("key1", 600, fn, allTrue)
	assert.Nil(t, err)
	_, err = rc.Fetch("key1", 600, errFn, allTrue)
	if expire == 0 {
		assert.ErrorIs(t, err, fetchError)
	} else {
		assert.Nil(t, err)
	}

	rc.Options.StrongConsistency = true
	_, err = rc.Fetch("key2", 600, fn, allTrue)
	assert.Nil(t, err)
	_, err = rc.Fetch("key2", 600, errFn, allTrue)
	if expire == 0 {
		assert.ErrorIs(t, err, fetchError)
	} else {
		assert.Nil(t, err)
	}
}

func TestErrorFetch(t *testing.T) {
	fn := func() (string, error) { return "", fmt.Errorf("error") }
	clearCache()
	rc := NewClient(rdb, NewDefaultOptions())
	_, err := rc.Fetch("key1", 60, fn, allTrue)
	assert.Equal(t, fmt.Errorf("error"), err)

	rc.Options.StrongConsistency = true
	_, err = rc.Fetch("key2", 60, fn, allTrue)
	assert.Equal(t, fmt.Errorf("error"), err)
}

func TestPanicFetch(t *testing.T) {
	fn := func() (string, error) { return "abc", nil }
	pfn := func() (string, error) { panic(fmt.Errorf("error")) }
	clearCache()
	rc := NewClient(rdb, NewDefaultOptions())
	_, err := rc.Fetch("key1", 60*time.Second, fn, allTrue)
	assert.Nil(t, err)
	rc.TagAsDeleted("key1")
	_, err = rc.Fetch("key1", 60*time.Second, pfn, allTrue)
	assert.Nil(t, err)
	time.Sleep(20 * time.Millisecond)
}

func TestTagAsDeletedWait(t *testing.T) {
	clearCache()
	rc := NewClient(rdb, NewDefaultOptions())
	rc.Options.WaitReplicas = 1
	rc.Options.WaitReplicasTimeout = 10
	err := rc.TagAsDeleted("key1")
	if getCluster() != nil {
		assert.Nil(t, err)
	} else {
		assert.Error(t, err, fmt.Errorf("wait replicas 1 failed. result replicas: 0"))
	}
}
