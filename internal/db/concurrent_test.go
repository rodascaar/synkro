package db

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestConcurrentWrites(t *testing.T) {
	tmp := t.TempDir()
	path := tmp + "/test.db"

	d, err := New(path)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()

	var wg sync.WaitGroup
	var mu sync.Mutex
	success, failures := 0, 0

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_, err := d.db.ExecContext(ctx, "INSERT INTO memories (id, created_at, updated_at, type, title, content, status) VALUES (?, datetime('now'), datetime('now'), 'note', ?, 'test', 'active')",
				fmt.Sprintf("test-%d", n), fmt.Sprintf("Test %d", n))
			mu.Lock()
			if err != nil {
				t.Errorf("goroutine %d failed: %v", n, err)
				failures++
			} else {
				success++
			}
			mu.Unlock()
		}(i)
	}

	wg.Wait()
	t.Logf("Result: %d success, %d failures", success, failures)
	if failures > 0 {
		t.Fail()
	}
}

func TestBusyTimeoutSet(t *testing.T) {
	tmp := t.TempDir()
	path := tmp + "/test.db"

	d, err := New(path)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()

	var bt int
	err = d.db.QueryRow("PRAGMA busy_timeout").Scan(&bt)
	if err != nil {
		t.Fatal(err)
	}
	if bt != 5000 {
		t.Errorf("expected busy_timeout=5000, got %d", bt)
	}
}
