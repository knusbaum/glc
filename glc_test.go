package glc

import (
	"context"
	"fmt"
	"io"
	"sync"
	"testing"
)

func TestEncstart(t *testing.T) {
	encstart(0x00112233, func() {
		id, ok := lastID()
		if !ok {
			t.Fail()
		}
		if id != 0x00112233 {
			t.Fail()
		}
	})
}

func TestContext(t *testing.T) {
	WithContext(context.WithValue(context.Background(), "foo", "bar"), func() {
		ctx := GetContext()
		v := ctx.Value("foo")
		if v != "bar" {
			t.Fail()
		}
	})
}

func BenchmarkWithContext(b *testing.B) {
	for i := 0; i < b.N; i++ {
		WithContext(context.WithValue(context.Background(), "foo", "bar"), func() {

		})
	}
}

func BenchmarkGetContext(b *testing.B) {
	WithContext(context.WithValue(context.Background(), "foo", "bar"), func() {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			GetContext()
			// v := ctx.Value("foo")
			// if v != "bar" {
			// 	b.Fail()
			// }

		}
	})
}

func BenchmarkWithContextGetContext(b *testing.B) {
	for i := 0; i < b.N; i++ {
		WithContext(context.WithValue(context.Background(), "foo", "bar"), func() {
			GetContext()
			// v := ctx.Value("foo")
			// if v != "bar" {
			// 	b.Fail()
			// }
		})
	}
}

func BenchmarkContention(b *testing.B) {
	var start, wg sync.WaitGroup
	routines := 200
	wg.Add(routines)
	start.Add(1)
	for i := 0; i < routines; i++ {
		go func() {
			start.Wait()
			defer wg.Done()
			WithContext(context.WithValue(context.Background(), "foo", "bar"), func() {
				for i := 0; i < b.N; i++ {
					GetContext()
					// v := ctx.Value("foo")
					// if v != "bar" {
					// 	b.Fail()
					// }

				}
			})
		}()
	}
	b.ResetTimer()
	start.Done()
	wg.Wait()
}

//go:noinline
func stackit(n int, f func()) {
	if n > 0 {
		stackit(n-1, f)
		return
	}
	f()
}

//go:noinline
func stackit2(n int, f func()) {
	if n > 0 {
		stackit2(n-1, f)
		return
	}
	f()
}

func OldBenchmarkDeepStack(b *testing.B) {
	var x int

	b.ResetTimer()
	WithContext(context.WithValue(context.Background(), "foo", "bar"), func() {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// This benchmark is dominated by the cost of stackit.
			//
			// We do one stackit that does nothing to eliminate the stackit portion
			// from the benchmark time.
			//
			// Exchanging the second stackit's n value with the first
			// will demonstrate the actual cost of a deep stack on GetContext.
			stackit2(1000, func() {
				x = 1
			})
			stackit(1000, func() {

				GetContext()
			})
		}
	})
	fmt.Fprintf(io.Discard, "X: %d\n", x)
}

func BenchmarkDeepStack(b *testing.B) {
	var x int
	WithContext(context.WithValue(context.Background(), "foo", "bar"), func() {
		stackit(1000, func() {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				GetContext()
			}
		})
	})
	fmt.Fprintf(io.Discard, "X: %d\n", x)
}
