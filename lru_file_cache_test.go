package goutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestLRUFileCache(t *testing.T) {
	data := bytes.Repeat([]byte{'a'}, 200)
	for i := 0; i < 100; i++ {
		fn := fmt.Sprintf("/tmp/LRUTEST%d", i)
		os.WriteFile(fn, data, 0666)
	}

	cache := NewLRUFileCache(1024, 10240)
	for i := 0; i < 100; i++ {
		fn := fmt.Sprintf("/tmp/LRUTEST%d", i)
		f, err := cache.Get(fn, time.Time{})
		if err != nil {
			t.Fatalf(err.Error())
		}
		if !bytes.Equal(f.Data, data) {
			t.Fatalf("data not equal")
		}
	}
	for i := 0; i < 100; i++ {
		fn := fmt.Sprintf("/tmp/LRUTEST%d", i)
		f, err := cache.Get(fn, time.Time{})
		if err != nil {
			t.Fatalf(err.Error())
		}
		if !bytes.Equal(f.Data, data) {
			t.Fatalf("data not equal")
		}
	}
	data = bytes.Repeat(data, 100)
	fn := "/tmp/fileSizeExceed"
	os.WriteFile(fn, data, 0666)
	cache.Clear()
	if _, err := cache.Get(fn, time.Time{}); err == nil {
		t.Fatal("file size limit test fail")
	} else {
		fmt.Println(err.Error())
	}

	s := cache.GetStatis()
	bts, _ := json.Marshal(s)
	fmt.Println(UnsafeBytesToString(bts))
	fmt.Println("TestLRUFileCache ok")
}
