package tools

import (
	"context"
	"testing"
)

func TestDeferredMemoryRunsOnlyWhenFlushedOutsideTransaction(t *testing.T) {
	type key struct{}
	parent := context.Background()
	work, flush := WithDeferredMemory(context.WithValue(parent, key{}, "transaction"))
	calls := 0
	if err := runMemoryEffect(work, func(ctx context.Context) error {
		calls++
		if ctx.Value(key{}) != nil {
			t.Fatal("embedding retained transaction")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if calls != 0 {
		t.Fatal("embedding executed before commit")
	}
	flush(parent)
	if calls != 1 {
		t.Fatal("embedding did not run after commit")
	}
	abandoned, _ := WithDeferredMemory(parent)
	_ = runMemoryEffect(abandoned, func(context.Context) error { t.Fatal("rolled-back embedding ran"); return nil })
}
