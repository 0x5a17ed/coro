package benchmarks

import (
	"testing"

	"github.com/0x5a17ed/coro"
)

func setupCoro() func(any) (int, bool) {
	return coro.NewSub(func(_ any, yield func(int) any) {
		for i := 0; ; i++ {
			yield(i)
		}
	}).Resume
}

func BenchmarkCoro(b *testing.B) {
	resume := setupCoro()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, _ = resume(nil)
	}
}
