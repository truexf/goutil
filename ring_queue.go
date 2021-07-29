// Copyright 2021 fangyousong(方友松). All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package goutil

import (
	"fmt"
	"sync"
	"time"
)

// 环形双向队列
type RingQueue struct {
	queue      []interface{}
	headPos    int
	tailPos    int
	capacity   int
	queueSize  int
	lock       sync.Mutex
	threadSafe bool
	queueSem   chan byte
}

func NewRingQueue(capacity int, threadSafe bool) *RingQueue {
	ret := new(RingQueue)
	ret.queue = make([]interface{}, capacity)
	ret.capacity = capacity
	ret.tailPos = capacity - 1
	ret.queueSem = make(chan byte, capacity)
	ret.threadSafe = threadSafe
	return ret
}

func (m *RingQueue) internalLock() {
	if m.threadSafe {
		m.lock.Lock()
	}
}

func (m *RingQueue) internalUnlock() {
	if m.threadSafe {
		m.lock.Unlock()
	}
}

func (m *RingQueue) Size() int {
	return m.queueSize
}

func (m *RingQueue) PushHead(value interface{}, waitTime time.Duration) error {
	getIt := false
	select {
	case m.queueSem <- 1:
		getIt = true
	default:
		getIt = false
	}
	if !getIt && waitTime > 0 {
		select {
		case m.queueSem <- 1:
			getIt = true
		case <-time.After(waitTime):
		}
	}
	if !getIt {
		return fmt.Errorf("queue full")
	}

	m.internalLock()
	m.queue[m.headPos] = value
	m.headPos++
	if m.headPos >= len(m.queue) {
		m.headPos = 0
	}
	m.queueSize++
	m.internalUnlock()

	return nil
}

func (m *RingQueue) PushTail(value interface{}, waitTime time.Duration) error {
	getIt := false
	select {
	case m.queueSem <- 1:
		getIt = true
	default:
		getIt = false
	}
	if !getIt && waitTime > 0 {
		select {
		case m.queueSem <- 1:
			getIt = true
		case <-time.After(waitTime):
		}
	}
	if !getIt {
		return fmt.Errorf("queue full")
	}

	m.internalLock()
	m.queue[m.tailPos] = value
	m.tailPos--
	if m.tailPos < 0 {
		m.tailPos = m.capacity - 1
	}
	m.queueSize++
	m.internalUnlock()

	return nil
}

func (m *RingQueue) PopHead(waitTime time.Duration) (interface{}, error) {
	getIt := false
	select {
	case <-m.queueSem:
		getIt = true
	default:
		getIt = false
	}
	if !getIt && waitTime > 0 {
		select {
		case <-m.queueSem:
			getIt = true
		case <-time.After(waitTime):
		}
	}
	if !getIt {
		return nil, fmt.Errorf("queue empty")
	}

	m.internalLock()
	headPos := m.headPos - 1
	if headPos < 0 {
		headPos = m.capacity - 1
	}
	ret := m.queue[headPos]
	m.headPos = headPos
	m.queueSize--
	m.internalUnlock()

	return ret, nil
}

func (m *RingQueue) PopTail(waitTime time.Duration) (interface{}, error) {
	getIt := false
	select {
	case <-m.queueSem:
		getIt = true
	default:
		getIt = false
	}
	if !getIt && waitTime > 0 {
		select {
		case <-m.queueSem:
			getIt = true
		case <-time.After(waitTime):
		}
	}
	if !getIt {
		return nil, fmt.Errorf("queue empty")
	}

	m.internalLock()
	pos := m.tailPos + 1
	if pos >= m.capacity {
		pos = 0
	}
	ret := m.queue[pos]
	m.tailPos = pos
	m.queueSize--
	m.internalUnlock()

	return ret, nil
}
