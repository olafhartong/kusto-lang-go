package kustolang

import "testing"

type pooledValue struct {
	id    int
	value string
}

func TestObjectPoolAllocatesCreatesAndReusesReturnedItem(t *testing.T) {
	nextID := 0
	resetCount := 0
	pool := NewObjectPool(func() *pooledValue {
		nextID++
		return &pooledValue{id: nextID}
	}, func(item *pooledValue) {
		resetCount++
		item.value = ""
	}, 2)

	first := pool.AllocateFromPool()
	first.value = "dirty"
	if first.id != 1 {
		t.Fatalf("first allocated id = %d, want 1", first.id)
	}

	pool.ReturnToPool(first)
	if resetCount != 1 {
		t.Fatalf("reset count = %d, want 1", resetCount)
	}
	if first.value != "" {
		t.Fatalf("returned item was not reset: %q", first.value)
	}

	second := pool.AllocateFromPool()
	if second != first {
		t.Fatal("expected pool to return the previously returned item")
	}
	if nextID != 1 {
		t.Fatalf("creator calls = %d, want 1", nextID)
	}
}

func TestObjectPoolDropsItemsWhenFull(t *testing.T) {
	nextID := 0
	pool := NewObjectPool(func() *pooledValue {
		nextID++
		return &pooledValue{id: nextID}
	}, func(item *pooledValue) {
		item.value = ""
	}, 1)

	first := pool.AllocateFromPool()
	second := pool.AllocateFromPool()
	pool.ReturnToPool(first)
	pool.ReturnToPool(second)

	reused := pool.AllocateFromPool()
	if reused != first {
		t.Fatalf("reused id = %d, want first id %d", reused.id, first.id)
	}

	newItem := pool.AllocateFromPool()
	if newItem == second {
		t.Fatal("extra returned item should have been dropped when the pool was full")
	}
	if newItem.id != 3 {
		t.Fatalf("new item id = %d, want 3", newItem.id)
	}
}

func TestObjectPoolDefaultSize(t *testing.T) {
	pool := NewObjectPool(func() *pooledValue {
		return &pooledValue{}
	}, func(item *pooledValue) {})
	if len(pool.items) != 10 {
		t.Fatalf("default pool size = %d, want 10", len(pool.items))
	}
}

func TestObjectPoolDoesNotStoreNilReference(t *testing.T) {
	created := 0
	reset := 0
	pool := NewObjectPool(func() *pooledValue {
		created++
		return &pooledValue{}
	}, func(*pooledValue) {
		reset++
	}, 1)

	pool.ReturnToPool(nil)
	if value := pool.AllocateFromPool(); value == nil {
		t.Fatal("nil return was stored in the pool")
	}
	if created != 1 || reset != 1 {
		t.Fatalf("created = %d, reset = %d", created, reset)
	}
}
