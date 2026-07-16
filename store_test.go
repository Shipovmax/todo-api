package main

import (
	"sync"
	"testing"
)

// TestStore_ConcurrentCreate verifies that concurrent writers never
// receive duplicate IDs and that every created task is retrievable
// afterwards. Run with -race to confirm the RWMutex actually guards
// concurrent access to the underlying map.
func TestStore_ConcurrentCreate(t *testing.T) {
	s := NewStore()

	const goroutines = 50
	ids := make(chan int, goroutines)

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			todo := s.Create("task")
			ids <- todo.ID
		}()
	}
	wg.Wait()
	close(ids)

	seen := make(map[int]bool, goroutines)
	for id := range ids {
		if seen[id] {
			t.Fatalf("duplicate ID assigned: %d", id)
		}
		seen[id] = true
	}
	if len(seen) != goroutines {
		t.Fatalf("expected %d unique IDs, got %d", goroutines, len(seen))
	}

	all := s.GetAll()
	if len(all) != goroutines {
		t.Fatalf("expected %d stored tasks, got %d", goroutines, len(all))
	}
}

// TestStore_CRUD exercises the basic Create/GetByID/Update/Delete flow.
func TestStore_CRUD(t *testing.T) {
	s := NewStore()

	todo := s.Create("write tests")
	if todo.ID == 0 {
		t.Fatalf("expected non-zero ID")
	}
	if todo.Done {
		t.Fatalf("expected new task to be not done")
	}

	got, ok := s.GetByID(todo.ID)
	if !ok || got.Title != "write tests" {
		t.Fatalf("GetByID(%d) = %+v, %v; want title %q, ok=true", todo.ID, got, ok, "write tests")
	}

	updated, ok := s.Update(todo.ID, true)
	if !ok || !updated.Done {
		t.Fatalf("Update(%d, true) = %+v, %v; want Done=true, ok=true", todo.ID, updated, ok)
	}

	if !s.Delete(todo.ID) {
		t.Fatalf("Delete(%d) = false; want true", todo.ID)
	}
	if _, ok := s.GetByID(todo.ID); ok {
		t.Fatalf("GetByID(%d) after delete: ok=true; want false", todo.ID)
	}
	if s.Delete(todo.ID) {
		t.Fatalf("Delete(%d) on already-deleted task = true; want false", todo.ID)
	}
}
