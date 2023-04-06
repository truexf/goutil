// 线程安全、持久化的map
// 包含一个持久化的map和一个先进先出的队列
// fangyousong 2023/3/17
package goutil

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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

func NewFileMap(tag string, filePath string, enableKeyQueue bool) (*FileMap, error) {
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
	}
	if enableKeyQueue {
		ret.keyQueue = make(chan string, 10000000)
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
	if m.fdDeleted != nil {
		m.fdDeleted.Sync()
	}
}
func (m *FileMap) Len() int {
	m.RLock()
	defer m.RUnlock()
	return len(m.dataMap)
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

func (m *FileMap) Put(key string, packedJsonData []byte, overrideIfExists bool, liveDuration time.Duration) error {
	if key == "" || len(packedJsonData) == 0 {
		return fmt.Errorf("invalid key or data")
	}
	if liveDuration > time.Hour*24*7 {
		return fmt.Errorf("live time too long")
	}
	if liveDuration <= 0 {
		return fmt.Errorf("live time too long")
	}

	putData := append(packedJsonData, '|')
	deadLine := fmt.Sprintf("%d", time.Now().Add(liveDuration).Unix())
	putData = append(putData, UnsafeStringToBytes(deadLine)...)

	m.Lock()
	defer m.Unlock()

	// write map
	_, exists := m.dataMap[key]
	if exists {
		if overrideIfExists {
			m.dataMap[key] = putData
		} else {
			return fmt.Errorf("key exists")
		}
	} else {
		m.dataMap[key] = putData
	}

	// write key queue
	if m.keyQueue != nil {
		m.keyQueue <- key
	}

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
	n, _ := m.fd.Write(UnsafeStringToBytes(key))
	m.fdSize += int64(n)
	m.fd.Write([]byte{'\000'})
	m.fdSize++
	writedBytes := 0
	for {
		n, err := m.fd.Write(putData[writedBytes:])
		if n > 0 {
			writedBytes += n
			m.fdSize += int64(n)
		}
		if err != nil || len(putData) == writedBytes {
			break
		}
	}
	m.fd.Write([]byte{'\n'})
	m.fdSize++

	return nil
}

// FIFO queue
func (m *FileMap) Pop(waitTimeout time.Duration) (retKey string, timeout bool) {
	if m.keyQueue == nil {
		return "", false
	}
	m.RLock()
	defer m.RUnlock()

	if waitTimeout <= 0 {
		select {
		case ret := <-m.keyQueue:
			return ret, false
		default:
			return "", true
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

func (m *FileMap) RemoveExpiredData() {
	m.Lock()
	defer m.Unlock()

	toDelete := make([]string, 0, 1024)
	tmNow := time.Now()
	for k, v := range m.dataMap {
		idx := bytes.LastIndexByte(v, '|')
		if idx >= 0 {
			u := string(v[idx+1:])
			ut, err := strconv.Atoi(u)
			if err != nil {
				toDelete = append(toDelete, k)
				continue
			}
			tmDeadLine := time.Unix(int64(ut), 0)
			if tmNow.After(tmDeadLine) {
				toDelete = append(toDelete, k)
			}
		}
	}
	for _, k := range toDelete {
		m.deleteNoLock(k)
	}
	if len(m.dataMap) == 0 {
		m.clear()
	}
}

func (m *FileMap) deleteNoLock(key string) {
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
	n, _ := m.fdDeleted.Write(UnsafeStringToBytes(key))
	if n > 0 {
		m.fdDeletedSize += int64(n)
	}
	m.fdDeleted.Write([]byte{'\n'})
	m.fdDeletedSize++

	if len(m.dataMap) == 0 {
		m.clear()
	}
}

func (m *FileMap) Delete(key string) {
	if key == "" {
		return
	}
	m.Lock()
	defer m.Unlock()
	m.deleteNoLock(key)
}

func (m *FileMap) Get(key string) []byte {
	if key == "" {
		return nil
	}

	m.RLock()
	defer m.RUnlock()

	if ret, ok := m.dataMap[key]; ok {
		idx := bytes.LastIndexByte(ret, '|')
		if idx >= 0 {
			ret = ret[:idx]
		}
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
					if m.keyQueue != nil {
						m.keyQueue <- key
					}
				}
			}
			if err != nil {
				break
			}
		}
		fd.Close()
	}
}

func (m *FileMap) clearFiles() {
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

func (m *FileMap) clear() {
	// close all fd
	m.fd.Close()
	m.fd = nil
	m.fdSize = 0
	m.fdDeleted.Close()
	m.fdDeleted = nil
	m.fdDeletedSize = 0

	// clear files
	m.clearFiles()
}
