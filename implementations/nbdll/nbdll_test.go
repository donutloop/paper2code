package nbdll

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
)

// -----------------------------------------------------------------------
// Basic sequential tests
// -----------------------------------------------------------------------

func TestEmptyList(t *testing.T) {
	l := NewList()
	if l.Len() != 0 {
		t.Fatalf("expected empty list, got len=%d", l.Len())
	}
}

func TestInsertBeforeSequential(t *testing.T) {
	l := NewList()
	c := &Cursor{}
	l.InitCursor(c)

	// Insert three items at EOL: [1] → [1,2] → [1,2,3]
	for _, v := range []int{1, 2, 3} {
		res := l.InsertBefore(c, v)
		if res != OK {
			t.Fatalf("InsertBefore(%d) = %v, want OK", v, res)
		}
	}
	snap := l.Snapshot()
	want := []interface{}{1, 2, 3}
	if !equal(snap, want) {
		t.Fatalf("snapshot = %v, want %v", snap, want)
	}
}

func TestDeleteSequential(t *testing.T) {
	l := NewList()
	c := &Cursor{}
	l.InitCursor(c)

	for _, v := range []int{10, 20, 30} {
		l.InsertBefore(c, v)
	}

	// Reset to start, then delete each element
	l.ResetCursor(c)
	for i := 0; i < 3; i++ {
		res := l.Delete(c)
		if res != OK {
			t.Fatalf("delete %d: got %v", i, res)
		}
	}
	// Now cursor is at EOL; deleting EOL returns false
	res := l.Delete(c)
	if res != ReturnFalse {
		t.Fatalf("delete at EOL: got %v, want ReturnFalse", res)
	}
	if l.Len() != 0 {
		t.Fatalf("list should be empty, len=%d", l.Len())
	}
}

func TestMoveRightLeft(t *testing.T) {
	l := NewList()
	c := &Cursor{}
	l.InitCursor(c)

	for _, v := range []int{1, 2, 3} {
		l.InsertBefore(c, v)
	}

	l.ResetCursor(c)

	// Walk right and collect values
	var got []interface{}
	for {
		v, res := l.Get(c)
		if res != OK {
			t.Fatalf("Get: %v", res)
		}
		if v == EOL {
			break
		}
		got = append(got, v)
		r, _ := l.MoveRight(c)
		if r == InvalidCursor {
			t.Fatal("unexpected InvalidCursor on MoveRight")
		}
	}
	want := []interface{}{1, 2, 3}
	if !equal(got, want) {
		t.Fatalf("right walk = %v, want %v", got, want)
	}

	// Walk left from EOL, but we need to step back once first (EOL is last real node)
	// We are at EOL; moveLeft takes us to 3
	var rev []interface{}
	for {
		r, ok := l.MoveLeft(c)
		if r == ReturnFalse || !ok {
			break
		}
		v, _ := l.Get(c)
		rev = append(rev, v)
	}
	wantRev := []interface{}{3, 2, 1}
	if !equal(rev, wantRev) {
		t.Fatalf("left walk = %v, want %v", rev, wantRev)
	}
}

func TestMoveRightAtEOL(t *testing.T) {
	l := NewList()
	c := &Cursor{}
	l.InitCursor(c)
	// Cursor starts at EOL in an empty list
	r, ok := l.MoveRight(c)
	if r != ReturnFalse || ok {
		t.Fatalf("MoveRight at EOL: got (%v,%v), want (ReturnFalse,false)", r, ok)
	}
}

func TestMoveLeftAtFirst(t *testing.T) {
	l := NewList()
	c := &Cursor{}
	l.InitCursor(c)
	// c is at EOL (first item in empty list)
	r, ok := l.MoveLeft(c)
	if r != ReturnFalse || ok {
		t.Fatalf("MoveLeft at first: got (%v,%v), want (ReturnFalse,false)", r, ok)
	}
}

func TestInsertBeforeMiddle(t *testing.T) {
	l := NewList()
	c1 := &Cursor{}
	c2 := &Cursor{}
	l.InitCursor(c1)
	l.InitCursor(c2)

	// Build [1, 3] via c1, then insert 2 between 1 and 3 using c2
	l.InsertBefore(c1, 1) // list: [1]
	l.InsertBefore(c1, 3) // list: [1, 3]; c1 is at EOL

	// Reset c2 and advance past 1 to point at 3
	l.ResetCursor(c2)
	l.MoveRight(c2) // c2 now at 3

	l.InsertBefore(c2, 2) // list: [1, 2, 3]

	snap := l.Snapshot()
	want := []interface{}{1, 2, 3}
	if !equal(snap, want) {
		t.Fatalf("snapshot = %v, want %v", snap, want)
	}
}

// -----------------------------------------------------------------------
// Concurrent stress tests
// -----------------------------------------------------------------------

// TestConcurrentInserts runs N goroutines each inserting M items.
// The list must contain N*M items at the end.
func TestConcurrentInserts(t *testing.T) {
	const (
		goroutines = 8
		perThread  = 200
	)
	l := NewList()
	var wg sync.WaitGroup

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			c := &Cursor{}
			l.InitCursor(c)
			for i := 0; i < perThread; i++ {
				for {
					res := l.InsertBefore(c, id*10000+i)
					if res == OK {
						break
					}
					// InvalidCursor: retry (cursor recovered internally)
				}
			}
			l.DestroyCursor(c)
		}(g)
	}
	wg.Wait()

	got := l.Len()
	want := goroutines * perThread
	if got != want {
		t.Fatalf("after concurrent inserts: len=%d, want %d", got, want)
	}
}

// TestConcurrentDeletes inserts items sequentially then deletes them
// concurrently from multiple goroutines.
func TestConcurrentDeletes(t *testing.T) {
	const total = 400
	l := NewList()

	// Insert sequentially
	c := &Cursor{}
	l.InitCursor(c)
	for i := 0; i < total; i++ {
		l.InsertBefore(c, i)
	}

	var deleted int64
	var wg sync.WaitGroup
	const goroutines = 4

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			dc := &Cursor{}
			l.InitCursor(dc)
			for {
				res := l.Delete(dc)
				switch res {
				case OK:
					atomic.AddInt64(&deleted, 1)
				case ReturnFalse:
					// At EOL – no more items
					l.DestroyCursor(dc)
					return
				case InvalidCursor:
					// Cursor was moved by a concurrent delete; retry
				}
			}
		}()
	}
	wg.Wait()

	if int(deleted) != total {
		t.Fatalf("deleted=%d, want %d", deleted, total)
	}
	if l.Len() != 0 {
		t.Fatalf("list not empty after all deletes, len=%d", l.Len())
	}
}

// TestConcurrentMixedOps runs concurrent inserts, deletes and moves.
// The test only checks that no goroutine panics or deadlocks.
func TestConcurrentMixedOps(t *testing.T) {
	const goroutines = 8
	const ops = 500

	l := NewList()

	// Pre-populate
	c0 := &Cursor{}
	l.InitCursor(c0)
	for i := 0; i < 50; i++ {
		l.InsertBefore(c0, i)
	}

	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			c := &Cursor{}
			l.InitCursor(c)
			rng := rand.New(rand.NewSource(int64(id)))
			for i := 0; i < ops; i++ {
				switch rng.Intn(4) {
				case 0:
					l.InsertBefore(c, rng.Int())
				case 1:
					l.Delete(c)
				case 2:
					l.MoveRight(c)
				case 3:
					l.MoveLeft(c)
				}
			}
			l.DestroyCursor(c)
		}(g)
	}
	wg.Wait()
}

// -----------------------------------------------------------------------
// Sorted-list maintenance (as described in §3 of the paper)
// -----------------------------------------------------------------------

// sortedInsert inserts v into list l while keeping it sorted, using the
// cursor-invalidation semantics described in §3 of the paper.
func sortedInsert(l *List, v int) {
	c := &Cursor{}
	l.InitCursor(c)
	defer l.DestroyCursor(c)

	for {
		val, res := l.Get(c)
		if res == InvalidCursor {
			// Cursor was moved by a concurrent insert; restart scan
			l.ResetCursor(c)
			continue
		}
		cur, isInt := val.(int)
		if val == EOL || (isInt && cur >= v) {
			// Insert before this position
			res2 := l.InsertBefore(c, v)
			if res2 == OK {
				return
			}
			// InvalidCursor → another value was inserted here; retry scan
			l.ResetCursor(c)
			continue
		}
		r, _ := l.MoveRight(c)
		if r == InvalidCursor {
			l.ResetCursor(c)
		}
	}
}

func isSorted(vals []interface{}) bool {
	for i := 1; i < len(vals); i++ {
		if vals[i-1].(int) > vals[i].(int) {
			return false
		}
	}
	return true
}

func TestConcurrentSortedInserts(t *testing.T) {
	const goroutines = 6
	const perThread = 40

	l := NewList()
	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			rng := rand.New(rand.NewSource(int64(id * 31)))
			for i := 0; i < perThread; i++ {
				sortedInsert(l, rng.Intn(1000))
			}
		}(g)
	}
	wg.Wait()

	snap := l.Snapshot()
	if len(snap) != goroutines*perThread {
		t.Fatalf("len=%d, want %d", len(snap), goroutines*perThread)
	}
	if !isSorted(snap) {
		t.Fatalf("list not sorted: %v", snap)
	}
}

// -----------------------------------------------------------------------
// Example
// -----------------------------------------------------------------------

func ExampleList() {
	l := NewList()
	c := &Cursor{}
	l.InitCursor(c)

	l.InsertBefore(c, "a")
	l.InsertBefore(c, "b")
	l.InsertBefore(c, "c")

	l.ResetCursor(c)
	for {
		v, _ := l.Get(c)
		if v == EOL {
			break
		}
		fmt.Print(v, " ")
		l.MoveRight(c)
	}
	fmt.Println()

	// Output: a b c
}

// -----------------------------------------------------------------------
// Helper
// -----------------------------------------------------------------------

func equal(a, b []interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
