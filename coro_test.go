// Copyright 2023 individual contributors. All rights reserved.
// Use of this source code is governed by a Zero-Clause BSD-style
// license that can be found in the LICENSE file.

package coro_test

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"

	"github.com/0x5a17ed/coro"
)

func consumeN[I, O any](cr *coro.C[I, O], t int) (out []O) {
	var zero I
	for n := 0; n < t; n++ {
		v, ok := cr.Resume(zero)
		if !ok {
			break
		}
		out = append(out, v)
	}
	return
}

func TestFlowFn(t *testing.T) {
	defer goleak.VerifyNone(t)

	log := make(chan string, 100)

	log <- "coro creation enter"
	cr := coro.NewFn(func(v string, yield func(int) string) int {
		log <- fmt.Sprint("generator enter s=", v)
		for i := 1; i < 4; i++ {
			log <- fmt.Sprint("generator send enter s=", i)
			v = yield(i)
			log <- fmt.Sprint("generator send leave s=", i, ",r=", v)
		}
		log <- fmt.Sprint("generator leave s=", 4)
		return 4
	})
	log <- "coro creation leave"
	defer cr.Stop()

	log <- "consuming enter"
	var received []int
	for _, s := range []string{"a", "b", "c", "d", "e"} {
		log <- fmt.Sprint("resume coro enter s=", s)
		v, ok := cr.Resume(s)
		log <- fmt.Sprint("resume coro leave s=", s, ",v=", v)
		if !ok {
			break
		}
		received = append(received, v)
	}
	log <- "consuming leave"
	close(log)

	var logLines []string
	for l := range log {
		logLines = append(logLines, l)
	}

	assert.Equal(t, []int{1, 2, 3, 4}, received)
	assert.Equal(t, []string{
		"coro creation enter",
		"coro creation leave",
		"consuming enter",
		"resume coro enter s=a",
		"generator enter s=a",
		"generator send enter s=1",
		"resume coro leave s=a,v=1",
		"resume coro enter s=b",
		"generator send leave s=1,r=b",
		"generator send enter s=2",
		"resume coro leave s=b,v=2",
		"resume coro enter s=c",
		"generator send leave s=2,r=c",
		"generator send enter s=3",
		"resume coro leave s=c,v=3",
		"resume coro enter s=d",
		"generator send leave s=3,r=d",
		"generator leave s=4",
		"resume coro leave s=d,v=4",
		"resume coro enter s=e",
		"resume coro leave s=e,v=0",
		"consuming leave",
	}, logLines)
}

func TestStop(t *testing.T) {
	t.Run("Twice", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		completedCh := make(chan struct{})

		go func() {
			defer close(completedCh)

			cr := coro.NewFn(func(_ struct{}, _ func(int) struct{}) int { return 0 })
			cr.Stop()
			cr.Stop()
		}()

		select {
		case <-completedCh:
		case <-time.After(5 * time.Second):
			assert.FailNowf(t, "timeout", "test timed out after %s", 5*time.Second)
		}
	})

	t.Run("Early", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		var lastValue int32

		cr := coro.NewSub(func(_ any, yield func(int32) any) {
			for {
				yield(atomic.AddInt32(&lastValue, 1))
			}
		})

		out := consumeN(cr, 4)
		cr.Stop()

		assert.Equal(t, []int32{1, 2, 3, 4}, out)
		assert.Equal(t, int32(4), lastValue)
	})
}

func TestPanicPropagation(t *testing.T) {
	tt := []struct {
		name string
		fn   func(cr *coro.C[any, int])
	}{
		{"Resume", func(cr *coro.C[any, int]) {
			cr.Resume(nil)
		}},
		{"Stop", func(cr *coro.C[any, int]) {
			cr.Stop()
		}},
	}
	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			defer goleak.VerifyNone(t)

			asserter := assert.New(t)

			cr := coro.NewSub(func(_ any, yield func(int) any) {
				defer func() {
					panic("yikes!")
				}()
				yield(13)
			})

			// Advancing the iterator the first time will yield the
			// value 13.
			var (
				yielded int
				ok      bool
			)
			asserter.NotPanics(func() {
				yielded, ok = cr.Resume(nil)
			})
			asserter.True(ok)
			asserter.Equal(13, yielded)

			// Advancing the iterator again should
			// crash the goroutine and the panic value
			// should be propagated.
			asserter.Panics(func() {
				tc.fn(cr)
			})
		})
	}
}
