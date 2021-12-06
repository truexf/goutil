// Copyright 2021 fangyousong(方友松). All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package goutil

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrorFileSizeLimited = errors.New("file size limited")
	ErrorFileNotExists   = errors.New("file not exists")
	ErrorFileNotModified = errors.New("file not modified")
)

type lruFileInfo struct {
	prior     *lruFileInfo
	next      *lruFileInfo
	sizeLimit int64
	loading   uint32
	lock      sync.RWMutex

	modifyTime time.Time
	loadTime   time.Time
	absPath    string
	data       []byte
	readCount  int64
	md5        string
}

type LRUFileStatis struct {
	AbsPath    string `json:"abs_path"`
	Size       int    `json:"size"`
	Md5        string `json:"md5"`
	ModifyTime int64  `json:"modify_time"`
	ReadCount  int64  `json:"read_count"`
}

type CachedFile struct {
	ModifyTime time.Time
	Data       []byte
	Md5        string
}

func (m *lruFileInfo) GetStatis() LRUFileStatis {
	return LRUFileStatis{
		AbsPath:    m.absPath,
		Size:       len(m.data),
		Md5:        m.md5,
		ModifyTime: m.modifyTime.Unix(),
		ReadCount:  m.readCount,
	}
}

func (m *lruFileInfo) load() (loading, modified bool, err error) {
	if atomic.CompareAndSwapUint32(&m.loading, 0, 1) {
		defer atomic.StoreUint32(&m.loading, 0)
		stat, err := os.Stat(m.absPath)
		if err != nil {
			return false, false, err
		}
		if stat.ModTime().Equal(m.modifyTime) {
			return false, false, nil
		}
		if stat.Size() > m.sizeLimit {
			return false, false, ErrorFileSizeLimited
		}
		if bts, err := ioutil.ReadFile(m.absPath); err != nil {
			return false, false, err
		} else {
			md5Hash := md5.New()
			if _, err := md5Hash.Write(bts); err != nil {
				return false, false, err
			}
			m.lock.Lock()
			defer m.lock.Unlock()
			m.md5 = fmt.Sprintf("%x", md5Hash.Sum(nil))
			m.data = bts
			m.modifyTime = stat.ModTime()
			m.loadTime = time.Now()
			return false, true, nil
		}
	} else {
		return true, false, nil
	}

}

type LRUFileCache struct {
	cacheLock      sync.RWMutex
	lruList        *lruFileInfo
	cacheMap       map[string]*lruFileInfo
	fileSizeLimit  int64
	cacheSizeLimit int64
	cacheSize      int64
}

type LRUFileCacheStatis struct {
	FileCount int             `json:"file_count"`
	CacheSize int64           `json:"cache_size"`
	FileList  []LRUFileStatis `json:"file_list"`
}

func NewLRUFileCache(fileSizeLimit int64, cacheSizeLimit int64) *LRUFileCache {
	ret := &LRUFileCache{
		cacheMap:       make(map[string]*lruFileInfo),
		lruList:        nil,
		fileSizeLimit:  fileSizeLimit,
		cacheSizeLimit: cacheSizeLimit,
	}
	if ret.cacheSizeLimit < ret.fileSizeLimit {
		ret.cacheSizeLimit = ret.fileSizeLimit * 1000
	}

	return ret
}

func (m *LRUFileCache) Clear() {
	m.cacheLock.Lock()
	defer m.cacheLock.Unlock()
	m.cacheMap = make(map[string]*lruFileInfo)
	m.lruList = nil
	m.cacheSize = 0
}

func (m *LRUFileCache) GetStatis() *LRUFileCacheStatis {
	m.cacheLock.RLock()
	defer m.cacheLock.RUnlock()
	ret := &LRUFileCacheStatis{FileCount: len(m.cacheMap), CacheSize: m.cacheSize, FileList: make([]LRUFileStatis, 0, len(m.cacheMap))}
	f := m.lruList
	for {
		if f != nil {
			ret.FileList = append(ret.FileList, f.GetStatis())
			f = f.next
		} else {
			break
		}
	}
	return ret
}

func (m *LRUFileCache) Get(abstractFilePath string, since time.Time) (*CachedFile, error) {
	m.cacheLock.RLock()
	f, exists := m.cacheMap[abstractFilePath]
	m.cacheLock.RUnlock()

	if !exists {
		f = &lruFileInfo{absPath: abstractFilePath, sizeLimit: m.fileSizeLimit}
	}

	var (
		loading, modified bool
		err               error
	)
	for {
		loading, modified, err = f.load()
		if err != nil {
			return nil, err
		}
		if !loading {
			break
		}
	}

	if !exists && !modified {
		return nil, ErrorFileNotExists
	}

	if !exists && modified {
		// new node append to header
		m.cacheLock.Lock()
		f.next = m.lruList
		if m.lruList != nil {
			m.lruList.prior = f
		}
		f.readCount++
		m.lruList = f
		m.cacheSize += int64(len(f.data))
		m.cacheMap[abstractFilePath] = f

		if m.cacheSize > m.cacheSizeLimit {
			// weed out
			tail := m.lruList
			for {
				if tail.next == nil {
					break
				}
				tail = tail.next
			}
			for {
				if m.cacheSize > m.cacheSizeLimit && tail.prior != nil {
					delete(m.cacheMap, tail.absPath)
					m.cacheSize -= int64(len(tail.data))
					tail = tail.prior
					tail.next = nil

				} else {
					break
				}
			}
		}
		m.cacheLock.Unlock()
	} else if f != m.lruList {
		// adjust f to header
		m.cacheLock.Lock()
		f.readCount++
		prior := f.prior
		next := f.next
		f.prior = nil
		f.next = m.lruList
		if prior != nil {
			prior.next = next
		}
		if next != nil {
			next.prior = prior
		}
		m.lruList = f
		m.cacheLock.Unlock()
	}

	if f.modifyTime.Before(since) {
		return nil, ErrorFileNotModified
	}
	f.lock.RLock()
	ret := &CachedFile{ModifyTime: f.modifyTime, Data: f.data, Md5: f.md5}
	f.lock.RUnlock()
	return ret, nil
}
