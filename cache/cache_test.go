package cache

import "testing"

func TestRedisCache_Has(t *testing.T) {
	err := testRedisCache.Forget("foo")
	if err != nil {
		t.Errorf("error forgetting key: %v", err)
	}

	exists, err := testRedisCache.Has("foo")
	if err != nil {
		t.Errorf("error checking key: %v", err)
	}

	if exists {
		t.Errorf("key should not exist")
	}

	testRedisCache.Set("foo", "bar", 60)
	exists, err = testRedisCache.Has("foo")
	if err != nil {
		t.Errorf("error checking key: %v", err)
	}

	if !exists {
		t.Errorf("key should exist")
	}
}

func TestRedisCache_Get(t *testing.T) {

	err := testRedisCache.Set("foo", "bar")
	if err != nil {
		t.Errorf("error setting key: %v", err)
	}

	value, err := testRedisCache.Get("foo")
	if err != nil {
		t.Errorf("error getting key: %v", err)
	}

	if value != "bar" {
		t.Errorf("value should be bar")
	}
}

func TestRedisCache_Forget(t *testing.T) {
	err := testRedisCache.Set("foo", "bar")
	if err != nil {
		t.Errorf("error setting key: %v", err)
	}

	err = testRedisCache.Forget("foo")
	if err != nil {
		t.Errorf("error forgetting key: %v", err)
	}

	exists, err := testRedisCache.Has("foo")
	if err != nil {
		t.Errorf("error checking key: %v", err)
	}

	if exists {
		t.Errorf("key should not exist")
	}
}

func TestRedisCache_Empty(t *testing.T) {
	err := testRedisCache.Set("alpha", "beta")
	if err != nil {
		t.Errorf("error setting key: %v", err)
	}

	err = testRedisCache.Empty()
	if err != nil {
		t.Errorf("error emptying cache: %v", err)
	}

	exists, err := testRedisCache.Has("alpha")
	if err != nil {
		t.Errorf("error checking key: %v", err)
	}

	if exists {
		t.Errorf("key should not exist")
	}

}

func TestRedisCache_EmptyByMatch(t *testing.T) {
	err := testRedisCache.Set("alpha", "foo")
	if err != nil {
		t.Errorf("error setting key: %v", err)
	}

	err = testRedisCache.Set("alpha2", "foo")
	if err != nil {
		t.Errorf("error setting key: %v", err)
	}

	err = testRedisCache.Set("beta", "bar")
	if err != nil {
		t.Errorf("error setting key: %v", err)
	}

	err = testRedisCache.EmptyByMatch("alpha")
	if err != nil {
		t.Errorf("error emptying cache: %v", err)
	}

	exists, err := testRedisCache.Has("alpha")
	if err != nil {
		t.Errorf("error checking key: %v", err)
	}

	if exists {
		t.Errorf("alpha found in cache, butkey should not exist")
	}

	exists, err = testRedisCache.Has("alpha2")
	if err != nil {
		t.Errorf("error checking key: %v", err)
	}

	if exists {
		t.Errorf("alpha2 found in cache, butkey should not exist")
	}

	exists, err = testRedisCache.Has("beta")
	if err != nil {
		t.Errorf("error checking key: %v", err)
	}

	if !exists {
		t.Errorf("beta not found in cache, butkey should exist")
	}

}
