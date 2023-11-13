// Copyright 2023 individual contributors. All rights reserved.
// Use of this source code is governed by a Zero-Clause BSD-style
// license that can be found in the LICENSE file.

// Package coro provides an implementation of coroutines built on
// top of Go's goroutines for the execution, suspension and resuming
// of generalized subroutines and functions for cooperative multitasking.
//
// Based on the wonderful [blog post] shared by Rus Cox.
//
// [blog post]: https://research.swtch.com/coro
package coro
