package goutil

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var MaxMapFileSize int64 = 1024 * 1024 * 1024

// 持久化的map
type FileMap struct {
	sync.RWMutex
	tag           string
	filePath      string
	dataMap       map[string][]byte
	keyQueue      chan string
	fd            *os.File
	fdSize        int64
	fdDeleted     *os.File
	fdDeletedSize int64
}

func NewFileMap(tag string, filePath string) (*FileMap, error) {
	if tag == "" || filePath == "" {
		return nil, fmt.Errorf("invalid param")
	}
	if fi, err := os.Stat(filePath); err != nil || !fi.IsDir() {
		return nil, fmt.Errorf("invalid param")
	}
	ret := &FileMap{
		tag:      tag,
		filePath: filePath,
		dataMap:  make(map[string][]byte),
		keyQueue: make(chan string, 10000000),
	}
	deleted := ret.loadDeletedKeysFromFile()
	ret.loadDataFromFile(deleted)
	return ret, nil
}

func (m *FileMap) Flush() {
	m.Lock()
	defer m.Unlock()

	if m.fd != nil {
		m.fd.Sync()
	}
}
func (m *FileMap) Close() {
	m.Lock()
	defer m.Unlock()

	m.fd.Close()
	m.fd = nil
	m.fdSize = 0
	m.fdDeleted.Close()
	m.fdDeleted = nil
	m.fdDeletedSize = 0
}

func (m *FileMap) getNextFileName(fileType string) (string, error) {
	i := 1
	fn := ""
	for {
		if i >= 999999 {
			return "", fmt.Errorf("file number for %s is too large.", m.tag)
		}
		fn = fmt.Sprintf("%s.%s.%s.%06d", m.tag, fileType, time.Now().Format("20060102"), i)
		if FileExists(filepath.Join(m.filePath, fn)) {
			i++
			continue
		}
		return filepath.Join(m.filePath, fn), nil
	}
}

func (m *FileMap) Put(key string, packedJsonData []byte, overrideIfExists bool) error {
	if key == "" || len(packedJsonData) == 0 {
		return fmt.Errorf("invalid key or data")
	}

	m.Lock()
	defer m.Unlock()

	// write map
	_, exists := m.dataMap[key]
	if exists {
		if overrideIfExists {
			m.dataMap[key] = packedJsonData
		} else {
			return fmt.Errorf("key exists")
		}
	} else {
		m.dataMap[key] = packedJsonData
	}

	// write key queue
	m.keyQueue <- key

	// write file
	if m.fd != nil {
		if m.fdSize >= MaxMapFileSize {
			m.fd.Close()
			m.fd = nil
			m.fdSize = 0
		}
	}
	if m.fd == nil {
		fn, err := m.getNextFileName("data")
		if err != nil {
			panic(err.Error())
		}
		m.fd, err = os.OpenFile(fn, os.O_APPEND|os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
		if err != nil {
			panic(err.Error())
		}
		m.fdSize = 0
	}
	n, _ := m.fd.Write([]byte(key))
	m.fdSize += int64(n)
	m.fd.Write([]byte("\000"))
	m.fdSize++
	writedBytes := 0
	for {
		n, err := m.fd.Write(packedJsonData[writedBytes:])
		if n > 0 {
			writedBytes += n
			m.fdSize += int64(n)
		}
		if err != nil || len(packedJsonData) == writedBytes {
			break
		}
	}
	m.fd.Write([]byte("\n"))
	m.fdSize++

	return nil
}

func (m *FileMap) Pop(waitTimeout time.Duration) (retKey string, timeout bool) {
	m.RLock()
	defer m.RUnlock()

	if waitTimeout <= 0 {
		select {
		case ret := <-m.keyQueue:
			return ret, false
		default:
			return "", false
		}
	} else {
		select {
		case ret := <-m.keyQueue:
			return ret, false
		case <-time.After(waitTimeout):
			return "", true
		}
	}
}

func (m *FileMap) Delete(key string) {
	if key == "" {
		return
	}
	m.Lock()
	defer m.Unlock()

	// delete from map
	delete(m.dataMap, key)

	//add delete record to file
	if m.fdDeleted != nil {
		if m.fdDeletedSize >= MaxMapFileSize {
			m.fdDeleted.Close()
			m.fdDeleted = nil
			m.fdDeletedSize = 0
		}
	}
	if m.fdDeleted == nil {
		fn, err := m.getNextFileName("deleted")
		if err != nil {
			panic(err.Error())
		}
		m.fdDeleted, err = os.OpenFile(fn, os.O_APPEND|os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
		if err != nil {
			panic(err.Error())
		}
		m.fdDeletedSize = 0
	}
	n, _ := m.fdDeleted.Write([]byte(key))
	if n > 0 {
		m.fdDeletedSize += int64(n)
	}
	m.fdDeleted.Write([]byte("\n"))
	m.fdDeletedSize++

	if len(m.dataMap) == 0 {
		m.clear()
	}
}

func (m *FileMap) Get(key string) []byte {
	if key == "" {
		return nil
	}

	m.RLock()
	defer m.RUnlock()

	if ret, ok := m.dataMap[key]; ok {
		return ret
	}
	return nil
}

func (m *FileMap) loadDeletedKeysFromFile() map[string]struct{} {
	m.Lock()
	defer m.Unlock()

	mpDeleted := make(map[string]struct{})
	des, err := os.ReadDir(m.filePath)
	if err != nil {
		panic(fmt.Sprintf("read dir: %s fail, %s", m.filePath, err.Error()))
	}
	for _, de := range des {
		if de.IsDir() {
			continue
		}
		if !strings.HasPrefix(de.Name(), fmt.Sprintf("%s.deleted", m.tag)) {
			continue
		}
		fd, err := os.Open(filepath.Join(m.filePath, de.Name()))
		if err != nil {
			fmt.Printf("open file %s fail, %s\n", de.Name(), err.Error())
			continue
		}
		bufReader := bufio.NewReaderSize(fd, 1024*1024*8)
		for {
			s, err := bufReader.ReadString('\n')
			if err == nil && len(s) > 0 && s[len(s)-1] == '\n' {
				s = s[:len(s)-1]
			}
			if len(s) > 0 {
				mpDeleted[s] = struct{}{}
			}
			if err != nil {
				break
			}
		}
		fd.Close()
	}

	return mpDeleted
}

func (m *FileMap) loadDataFromFile(deleted map[string]struct{}) {
	m.Lock()
	defer m.Unlock()

	des, err := os.ReadDir(m.filePath)
	if err != nil {
		panic(fmt.Sprintf("read dir: %s fail, %s", m.filePath, err.Error()))
	}
	for _, de := range des {
		if de.IsDir() {
			continue
		}
		if !strings.HasPrefix(de.Name(), fmt.Sprintf("%s.data", m.tag)) {
			continue
		}
		fd, err := os.Open(filepath.Join(m.filePath, de.Name()))
		if err != nil {
			fmt.Printf("open file %s fail, %s\n", de.Name(), err.Error())
			continue
		}
		bufReader := bufio.NewReaderSize(fd, 1024*1024*8)
		for {
			s, err := bufReader.ReadSlice('\n')
			if err == nil && len(s) > 0 && s[len(s)-1] == '\n' {
				s = s[:len(s)-1]
			}
			if len(s) > 0 {
				idx := bytes.IndexByte(s, '\000')
				if idx < 0 || idx >= len(s)-1 {
					continue
				}
				key := string(s[:idx])
				if key == "" {
					continue
				}
				data := s[idx+1:]
				if _, ok := deleted[key]; !ok {
					m.dataMap[key] = data
					m.keyQueue <- key
				}
			}
			if err != nil {
				break
			}
		}
		fd.Close()
	}
}

func (m *FileMap) clear() {
	// close all fd
	m.fd.Close()
	m.fd = nil
	m.fdSize = 0
	m.fdDeleted.Close()
	m.fdDeleted = nil
	m.fdDeletedSize = 0

	// clear files
	des, err := os.ReadDir(m.filePath)
	if err != nil {
		panic(fmt.Sprintf("read dir: %s fail, %s", m.filePath, err.Error()))
	}
	for _, de := range des {
		if de.IsDir() {
			continue
		}
		if !strings.HasPrefix(de.Name(), fmt.Sprintf("%s.data", m.tag)) &&
			!strings.HasPrefix(de.Name(), fmt.Sprintf("%s.deleted", m.tag)) {
			continue
		}
		os.Remove(filepath.Join(m.filePath, de.Name()))
	}
}
