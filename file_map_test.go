package goutil

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

var once sync.Once
var fileMapObj *FileMap
var err error

func ensureInit() {
	once.Do(func() {
		os.MkdirAll("/tmp/fm", 0777)
		fileMapObj, err = NewFileMap("testFileMap", "/tmp/fm", true)
		if err != nil {
			panic(err)
		}
	})
}

func TestFileMap_PutAndGet(t *testing.T) {
	ensureInit()
	for i := 1; i < 1000; i++ {
		if err := fileMapObj.Put("testkey"+strconv.Itoa(i), []byte(strings.Repeat("testdata"+strconv.Itoa(i), i)), true, time.Second*time.Duration(i)); err != nil {
			t.Fatal(err.Error())
		}
	}
	for i := 1; i < 1000; i++ {
		if string(fileMapObj.Get("testkey"+strconv.Itoa(i))) != strings.Repeat("testdata"+strconv.Itoa(i), i) {
			t.Fatal("get is not equal for put")
		}
	}
	fileMapObj.Close()
	fmt.Println("test TestFileMap_PutAndGet ok.")
}

func TestNewFileMap(t *testing.T) {
	fm, err := NewFileMap("testFileMap", "/tmp/fm", true)
	if err != nil {
		t.Fatal(err)
	}
	defer fm.Close()
	for i := 1; i < 1000; i++ {
		if string(fm.Get("testkey"+strconv.Itoa(i))) != strings.Repeat("testdata"+strconv.Itoa(i), i) {
			t.Fatal("get is not equal for put")
		}
	}

	fmt.Println("test TestNewFileMap ok.")
}

func TestFileMapExpire(t *testing.T) {
	time.Sleep(time.Second * 2)
	fm, err := NewFileMap("testFileMap", "/tmp/fm", false)
	if err != nil {
		t.Fatal(err)
	}
	defer fm.Close()
	fm.RemoveExpiredData()
	fmt.Printf("after expired, len: %d\n", fm.Len())
	//if fm.Len() > 995 {
	//	t.Fatal("testfilemapexpire fail")
	//}
	//fmt.Println("testfilemapexpire ok ")
}

func TestFileMap_Pop(t *testing.T) {
	fm, err := NewFileMap("testFileMap", "/tmp/fm", true)
	if err != nil {
		t.Fatal(err)
	}
	defer fm.Close()
	for i := 1; i < 1000; i++ {
		ret, _ := fm.Pop(time.Second)
		if ret != "testkey"+strconv.Itoa(i) {
			t.Fatal("pop data != push key")
		}
	}
	fmt.Println("test TestFileMap_Pop ok.")
}

func TestFileMap_Delete(t *testing.T) {
	fm, err := NewFileMap("testFileMap", "/tmp/fm", true)
	if err != nil {
		t.Fatal(err)
	}
	defer fm.Close()
	for i := 1; i < 1000; i++ {
		fm.Delete("testkey" + strconv.Itoa(i))
		if string(fm.Get("testkey"+strconv.Itoa(i))) == strings.Repeat("testdata"+strconv.Itoa(i), i) {
			t.Fatal("get ok after delete")
		}
	}
	fmt.Println("test TestFileMap_Delete ok.")
}

// final, clean
func TestFileMap_Clear(t *testing.T) {
	return
	fm, err := NewFileMap("testFileMap", "/tmp/fm", true)
	if err != nil {
		t.Fatal(err)
	}
	fm.clear()
	fm.Close()
}
