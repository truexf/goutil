// Copyright 2021 fangyousong(方友松). All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package goutil

import (
	"sync/atomic"
	"time"
)

// 令牌桶，用于限流.
// 令牌生产的时间间隔固定为100ms，每次生产的令牌数根据qps的大小计算出：
// qps/10+1
type TokenBucket struct {
	//当前令牌数
	tokenCount int64
	//令牌桶容量
	capacity   int64
	stopNotify chan int
	qpsChan    chan int64
}

func NewTokenBucket(capacity int64, qps int64) *TokenBucket {
	if qps < 100 {
		return nil
	}
	ret := &TokenBucket{capacity: capacity}
	ret.stopNotify = make(chan int, 1)
	ret.qpsChan = make(chan int64, 1)
	ret.qpsChan <- qps
	go ret.run()
	return ret
}

//启动
func (m *TokenBucket) run() {
	ticker := time.NewTicker(time.Millisecond * 100)
	qps := <-m.qpsChan
	rate := qps/10 + 1
	if m.capacity < rate {
		m.capacity = rate
	}
	m.tokenCount = rate
	for {
		select {
		case <-ticker.C:
			cnt := atomic.LoadInt64(&m.tokenCount) + rate
			if cnt > m.capacity {
				cnt = m.capacity
			}
			atomic.StoreInt64(&m.tokenCount, cnt)
		case qps := <-m.qpsChan:
			if qps < 100 {
				continue
			}
			rate = qps/10 + 1
			if m.capacity < rate {
				m.capacity = rate
			}
		case <-m.stopNotify:
			ticker.Stop()
			return
		}
	}
}

//修改qps
func (m *TokenBucket) SetQps(qps int64) {
	m.qpsChan <- qps
}

//停止生产令牌
func (m *TokenBucket) Stop() {
	m.stopNotify <- 1
}

//获取令牌
func (m *TokenBucket) GetToken() bool {
	if tokenCount := atomic.LoadInt64(&m.tokenCount); tokenCount > 0 {
		atomic.AddInt64(&m.tokenCount, -1)
		return true
	} else {
		return false
	}
}
