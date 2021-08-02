package goutil

import (
	"sync"
	"sync/atomic"
	"testing"
)

type TestData struct {
	v string
	n int
}

func TestLinkedList(t *testing.T) {
	ll := NewLinkedList(true)
	var wg sync.WaitGroup
	wg.Add(3)
	var pushCnt uint32
	var popCnt uint32
	for i := 0; i < 3; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 100000; j++ {
				ll.PushTail(&TestData{v: "hello", n: j}, true)
				atomic.AddUint32(&pushCnt, 1)
			}
		}()
	}
	wg.Wait()
	wg.Add(3)
	for i := 0; i < 3; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 100000; j++ {
				ll.PopHead(true)
				atomic.AddUint32(&popCnt, 1)
			}
		}()
	}
	wg.Wait()
	if ll.Len != 0 {
		t.Fatalf("fail, len: %d, push %d, pop %d\n", ll.Len, pushCnt, popCnt)
	}
}
