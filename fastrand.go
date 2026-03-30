package goutil

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

type LockedRand struct {
	pool sync.Pool
}

var (
	// 使用原子计数器配合时间戳，确保即便在高并发初始化时，每个实例的 Seed 也不一样
	seedCounter int64
)

func NewLockedRand() *LockedRand {
	return &LockedRand{
		pool: sync.Pool{
			New: func() interface{} {
				// 核心：为池中的每个实例生成唯一的 Seed
				seed := time.Now().UnixNano() + atomic.AddInt64(&seedCounter, 1)
				return rand.New(rand.NewSource(seed))
			},
		},
	}
}

// Intn 封装了从 Pool 取出和放回的过程
func (r *LockedRand) Intn(n int) int {
	if n <= 0 {
		return 0
	}
	// 从池中获取一个 *rand.Rand 实例
	rng := r.pool.Get().(*rand.Rand)
	res := rng.Intn(n)
	// 用完后放回，供其他 Goroutine 使用
	r.pool.Put(rng)
	return res
}

// 如果你需要全局单例
var globalFastRand = NewLockedRand()

func Intn(n int) int {
	return globalFastRand.Intn(n)
}
