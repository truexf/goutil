//线程安全的随机数生成器
package goutil

import (
	"math/rand"
	"sync"
	"time"
)

type SafeRand struct {
	sync.Mutex
	mathRand rand.Source
}

func NewSafeRand() *SafeRand {
	ret := new(SafeRand)
	ret.mathRand = rand.NewSource(time.Now().UnixNano())
	return ret
}

func (m *SafeRand) Int63n(n int64) int64 {
	m.Lock()
	defer m.Unlock()
	if n <= 0 {
		panic("invalid argument to Int63n")
	}
	if n&(n-1) == 0 { // n is power of two, can mask
		return m.mathRand.Int63() & (n - 1)
	}
	max := int64((1 << 63) - 1 - (1<<63)%uint64(n))
	v := m.mathRand.Int63()
	for v > max {
		v = m.mathRand.Int63()
	}
	return v % n
}
