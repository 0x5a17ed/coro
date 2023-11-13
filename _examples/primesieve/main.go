package main

import (
	"fmt"
	"runtime"
	"runtime/debug"

	"github.com/0x5a17ed/coro"
)

func counter(start int) *coro.C[any, int] {
	return coro.NewSub(func(_ any, yield func(int) any) {
		for i := start; ; i++ {
			yield(i)
		}
	})
}

func filter(p int, cr *coro.C[any, int]) *coro.C[any, int] {
	return coro.NewSub(func(_ any, yield func(int) any) {
		for {
			n, _ := cr.Resume(nil)
			if n%p != 0 {
				yield(n)
			}
		}
	})
}

func primes() *coro.C[any, int] {
	return coro.NewSub(func(_ any, yield func(int) any) {
		cr := counter(2)
		defer cr.Stop()

		for {
			p, _ := cr.Resume(nil)
			yield(p)

			cr = filter(p, cr)
			defer cr.Stop()
		}
	})
}

func takeN[T any](n int, cr *coro.C[any, T]) (out []T) {
	for i := 0; i < n; i++ {
		v, ok := cr.Resume(nil)
		if !ok {
			break
		}

		out = append(out, v)
	}
	return out
}

func printPrimes() {
	cr := primes()
	defer cr.Stop()

	fmt.Println(takeN(15, cr))
}

func main() {
	printPrimes()

	debug.SetTraceback("all")
	fmt.Println(runtime.NumGoroutine(), "goroutines")
}
