// Copyright 2021 fangyousong(方友松). All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package goutil

import (
	"fmt"
	"sync"
	"time"
)

// 由多个环形队列拼接成的长队列，借鉴了c++ stl中dequeue的思想
type Queue struct {
	chunkList *LinkedList
	chunkSize int
	capacity  int
	dataSem   chan byte
	lock      sync.Mutex
}

func NewQueue(capacity int, chunkSize int) *Queue {
	ret := &Queue{}
	ret.capacity = capacity
	ret.chunkSize = chunkSize
	ret.chunkList = NewLinkedList(false)
	ret.dataSem = make(chan byte, capacity)
	return ret
}

func (m *Queue) Produce(data interface{}, waitTime time.Duration) error {
	getIt := false
	tmr := AcquireTimer(waitTime)
	defer ReleaseTimer(tmr)
	select {
	case m.dataSem <- 1:
		getIt = true
	case <-tmr.C:
	}
	if !getIt {
		return fmt.Errorf("queue was full")
	}
	m.lock.Lock()
	defer m.lock.Unlock()
	var chunk *RingQueue = nil
	if m.chunkList.tail == nil {
		chunk = NewRingQueue(m.chunkSize, false)
		m.chunkList.PushTail(chunk, true)
	} else {
		chunk = m.chunkList.tail.Data.(*RingQueue)
	}
	if err := chunk.PushTail(data, 0); err != nil {
		//queue was full
		chunk = NewRingQueue(m.chunkSize, false)
		m.chunkList.PushTail(chunk, true)
		chunk.PushTail(data, 0)
	}
	return nil
}

func (m *Queue) Consume(waitTime time.Duration) (interface{}, error) {
	getIt := false
	tmr := AcquireTimer(waitTime)
	defer ReleaseTimer(tmr)
	select {
	case <-m.dataSem:
		getIt = true
	case <-tmr.C:
	}
	if !getIt {
		return nil, fmt.Errorf("queue was empty")
	}
	m.lock.Lock()
	defer m.lock.Unlock()
	var chunk *RingQueue = nil
	chunk = m.chunkList.head.Data.(*RingQueue)
	ret, _ := chunk.PopHead(0)
	if chunk.Size() == 0 {
		m.chunkList.PopHead(true)
	}
	return ret, nil
}
