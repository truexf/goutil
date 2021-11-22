// Copyright 2021 fangyousong(方友松). All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package goutil

import (
	"sync"
	"time"
)

func initTimer(t *time.Timer, timeout time.Duration) *time.Timer {
	if t == nil {
		return time.NewTimer(timeout)
	}
	if t.Reset(timeout) {
		panic("BUG: active timer trapped into initTimer()")
	}
	return t
}

func stopTimer(t *time.Timer) {
	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}
}

func AcquireTimer(timeout time.Duration) *time.Timer {
	v := timerPool.Get()
	if v == nil {
		return time.NewTimer(timeout)
	}
	t := v.(*time.Timer)
	initTimer(t, timeout)
	return t
}

func ReleaseTimer(t *time.Timer) {
	stopTimer(t)
	timerPool.Put(t)
}

var timerPool sync.Pool
