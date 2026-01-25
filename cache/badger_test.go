package cache

import "testing"

func TestBadgerCache_Has(t *testing.T) {
	err := testBadgerCache.Forget("foo")
	if err != nil {
		t.Error(err)
	}
	has, err := testBadgerCache.Has("foo")
	if err != nil {
		t.Error(err)
	}
	if has {
		t.Error("foo found in cache, should not be")
	}

	_ = testBadgerCache.Set("foo", "bar")
	has, err = testBadgerCache.Has("foo")
	if err != nil {
		t.Error(err)
	}
	if !has {
		t.Error("foo not found in cache, should be")
	}

	err = testBadgerCache.Forget("foo")

}

func TestBadgerCache_Get(t *testing.T) {
	err := testBadgerCache.Set("foo", "bar")
	if err != nil {
		t.Error(err)
	}

	val, err := testBadgerCache.Get("foo")
	if err != nil {
		t.Error(err)
	}

	if val != "bar" {
		t.Error("foo not found in cache, should be")
	}
}

func TestBadgerCache_Forget(t *testing.T) {
	err := testBadgerCache.Set("foo", "bar")
	if err != nil {
		t.Error(err)
	}

	err = testBadgerCache.Forget("foo")
	if err != nil {
		t.Error(err)
	}

	has, err := testBadgerCache.Has("foo")
	if err != nil {
		t.Error(err)
	}
	if has {
		t.Error("foo found in cache, should not be")
	}

}

func TestBadgerCache_Empty(t *testing.T) {
	err := testBadgerCache.Set("foo", "bar")
	if err != nil {
		t.Error(err)
	}

	err = testBadgerCache.Empty()
	if err != nil {
		t.Error(err)
	}

	has, err := testBadgerCache.Has("foo")
	if err != nil {
		t.Error(err)
	}
	if has {
		t.Error("foo found in cache, should not be")
	}
}

func TestBadgerCache_EmptyByMatch(t *testing.T) {
	err := testBadgerCache.Set("alpha", "bar")
	if err != nil {
		t.Error(err)
	}

	err = testBadgerCache.Set("alpha2", "bar")
	if err != nil {
		t.Error(err)
	}

	err = testBadgerCache.Set("beta", "bar")
	if err != nil {
		t.Error(err)
	}

	err = testBadgerCache.EmptyByMatch("a")
	if err != nil {
		t.Error(err)
	}

	has, err := testBadgerCache.Has("alpha")
	if err != nil {
		t.Error(err)
	}
	if has {
		t.Error("alpha found in cache, should not be")
	}

	has, err = testBadgerCache.Has("alpha2")
	if err != nil {
		t.Error(err)
	}
	if has {
		t.Error("alpha2 found in cache, should not be")
	}

	has, err = testBadgerCache.Has("beta")
	if err != nil {
		t.Error(err)
	}
	if !has {
		t.Error("beta not found in cache, should be")
	}
}
