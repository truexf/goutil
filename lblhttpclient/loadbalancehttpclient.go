// Copyright 2021 fangyousong(方友松). All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package lblhttpclient

import (
	"errors"
	"hash/fnv"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/truexf/goutil/jsonexp"
)

type LoadBalanceMethod uint32

const (
	MethodRoundrobin LoadBalanceMethod = iota
	MethodRandom
	MethodMinPending
	MethodIpHash
	MethodUrlParam
	MethodJsonExp // very powerful
)

const (
	DefaultConnectTimeout      time.Duration = time.Second * 3
	DefaultWaitResponseTimeout time.Duration = time.Second

	JsonExpGroupTarget     string = "LB_TARGET"
	JsonExpVarTargetServer string = "$LB_TARGET_SERVER"
	JsonExpUrlPath         string = "__PATH__"
	JsonExpObjectURI       string = "$REQUEST_URI"
)

var (
	ErrorAliasExists     = errors.New("backend's alias of load balance client exists")
	ErrorJsonExpNotFound = errors.New("jsonexp for target select not found")
	ErrorNoServerDefined = errors.New("no backend server defined")
	jsonExpDict          *jsonexp.Dictionary
)

type UrlValuesForJsonExp struct {
	Path      string
	UrlValues url.Values
}

// implements jsonexp.Object
func (m *UrlValuesForJsonExp) SetPropertyValue(property string, value interface{}, context jsonexp.Context) {
}

func (m *UrlValuesForJsonExp) GetPropertyValue(property string, context jsonexp.Context) interface{} {
	if property == JsonExpUrlPath {
		return m.Path
	}
	return m.UrlValues.Get(property)
}

type lblHttpBackend struct {
	pendingRequests int64
	httpClient      *http.Client
	addr            string // ip:port
	alias           string
}

func doInit() {
	var once sync.Once
	once.Do(func() {
		jsonExpDict = jsonexp.NewDictionary()
		jsonExpDict.RegisterVar(JsonExpVarTargetServer, nil)
	})
}

func newLblHttpBackend(addr string, alias string, maxIdleConns int, connTimeout, waitTimeout time.Duration) (*lblHttpBackend, error) {
	if _, err := net.ResolveTCPAddr("tcp4", addr); err != nil {
		return nil, err
	}
	ret := &lblHttpBackend{
		addr:  addr,
		alias: alias,
		httpClient: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   connTimeout,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				ForceAttemptHTTP2:     true,
				MaxIdleConns:          maxIdleConns,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
				ResponseHeaderTimeout: waitTimeout,
			},
		},
	}

	return ret, nil
}

type LblHttpClient struct {
	randObj           *rand.Rand
	serverListLock    sync.RWMutex
	serverList        []*lblHttpBackend
	serverMap         map[string]*lblHttpBackend // key is backend's alias
	method            LoadBalanceMethod
	roundrobinIndex   int64
	methodUrlParamKey string

	// method MethodJsonExp
	jsonExpLock   sync.RWMutex
	jsonExpConfig *jsonexp.Configuration

	maxIdleConnectionsPerServer int
	connectTimeout              time.Duration
	waitResponseTimeout         time.Duration
}

func NewLoadBalanceClient(method LoadBalanceMethod, maxIdleConnectionsPerServer int, methodUrlParamKey string, connTimeout, waitResponseTimeout time.Duration) *LblHttpClient {
	doInit()

	if connTimeout <= 0 {
		connTimeout = DefaultConnectTimeout
	}
	if waitResponseTimeout < 0 {
		waitResponseTimeout = DefaultWaitResponseTimeout
	}

	ret := &LblHttpClient{
		randObj:                     rand.New(rand.NewSource(time.Now().UnixNano())),
		method:                      method,
		methodUrlParamKey:           methodUrlParamKey,
		maxIdleConnectionsPerServer: maxIdleConnectionsPerServer,
		connectTimeout:              connTimeout,
		waitResponseTimeout:         waitResponseTimeout,
		serverList:                  make([]*lblHttpBackend, 0),
		serverMap:                   make(map[string]*lblHttpBackend),
	}

	return ret
}

func (m *LblHttpClient) SetJsonExp(jsonExpJson []byte) error {
	config, err := jsonexp.NewConfiguration(jsonExpJson, jsonExpDict)
	if err != nil {
		return err
	}
	m.jsonExpLock.Lock()
	defer m.jsonExpLock.Unlock()
	m.jsonExpConfig = config
	return nil
}

func (m *LblHttpClient) GetJsonExp() *jsonexp.JsonExpGroup {
	m.jsonExpLock.RLock()
	defer m.jsonExpLock.RUnlock()
	if m.jsonExpConfig == nil {
		return nil
	}
	if ret, ok := m.jsonExpConfig.GetJsonExpGroup(JsonExpGroupTarget); ok {
		return ret
	} else {
		return nil
	}
}

func (m *LblHttpClient) AddBackend(addr string, alias string) error {
	if _, ok := m.serverMap[alias]; ok {
		return ErrorAliasExists
	}
	backend, err := newLblHttpBackend(addr, alias, m.maxIdleConnectionsPerServer, m.connectTimeout, m.waitResponseTimeout)
	if err != nil {
		return err
	}

	m.serverListLock.Lock()
	defer m.serverListLock.Unlock()
	m.serverList = append(m.serverList, backend)
	m.serverMap[alias] = backend

	return nil
}

func (m *LblHttpClient) selectBackendRoundrobin() (*lblHttpBackend, error) {
	idx := atomic.LoadInt64(&m.roundrobinIndex)
	atomic.AddInt64(&m.roundrobinIndex, 1)
	m.serverListLock.RLock()
	idx %= int64(len(m.serverList))
	ret := m.serverList[idx]
	m.serverListLock.RUnlock()
	return ret, nil
}

func (m *LblHttpClient) selectBackendRandom() (*lblHttpBackend, error) {
	m.serverListLock.RLock()
	idx := m.randObj.Intn(len(m.serverList))
	ret := m.serverList[idx]
	m.serverListLock.RUnlock()
	return ret, nil
}

func (m *LblHttpClient) selectBackendMinPending() (*lblHttpBackend, error) {
	idx := atomic.LoadInt64(&m.roundrobinIndex)
	atomic.AddInt64(&m.roundrobinIndex, 1)
	m.serverListLock.RLock()
	idx %= int64(len(m.serverList))
	minIdx := idx
	minPending := m.serverList[minIdx].pendingRequests
	if minPending > 0 {
		for i := 0; i < len(m.serverList); i++ {
			idx++
			if idx >= int64(len(m.serverList)) {
				idx = 0
			}
			pr := m.serverList[idx].pendingRequests
			if pr < minPending {
				minPending = pr
				minIdx = idx
			}
		}
	}
	m.serverListLock.RUnlock()
	return m.serverList[minIdx], nil
}

func (m *LblHttpClient) selectBackendIpHash(clientIp string) (*lblHttpBackend, error) {
	fnv32 := fnv.New32()
	io.WriteString(fnv32, clientIp)
	idx := fnv32.Sum32()
	m.serverListLock.RLock()
	defer m.serverListLock.RUnlock()
	idx = idx % uint32(len(m.serverList))
	return m.serverList[int(idx)], nil
}

func (m *LblHttpClient) selectBackendUrlParam(paramValue string) (*lblHttpBackend, error) {
	fnv32 := fnv.New32()
	io.WriteString(fnv32, paramValue)
	idx := fnv32.Sum32()
	m.serverListLock.RLock()
	defer m.serverListLock.RUnlock()
	idx = idx % uint32(len(m.serverList))
	return m.serverList[int(idx)], nil
}

func (m *LblHttpClient) selectBackendJsonExp() (*lblHttpBackend, error) {
	jsonExp := m.GetJsonExp()
	if jsonExp == nil {
		return nil, ErrorJsonExpNotFound
	}

	context := &jsonexp.DefaultContext{WithLock: false}
	if err := jsonExp.Execute(context); err != nil {
		return nil, err
	}
	targetAlias, ok := context.GetCtxData(JsonExpVarTargetServer)
	if !ok {
		return nil, ErrorJsonExpNotFound
	}
	targetAliasStr, _ := jsonexp.GetStringValue(targetAlias)
	m.serverListLock.RLock()
	defer m.serverListLock.RUnlock()
	if backend, ok := m.serverMap[targetAliasStr]; ok {
		return backend, nil
	}
	return nil, ErrorJsonExpNotFound
}

func (m *LblHttpClient) selectBackend(clientIp string, request *http.Request) (*lblHttpBackend, error) {
	if len(m.serverList) == 0 {
		return nil, ErrorNoServerDefined
	}
	switch m.method {
	case MethodRoundrobin:
		return m.selectBackendRoundrobin()
	case MethodRandom:
		return m.selectBackendRandom()
	case MethodMinPending:
		return m.selectBackendMinPending()
	case MethodIpHash:
		return m.selectBackendIpHash(clientIp)
	case MethodUrlParam:
		u := request.URL
		paramValue := u.Query().Get(m.methodUrlParamKey)
		return m.selectBackendUrlParam(paramValue)
	case MethodJsonExp:
		u := request.URL
		jsonExpDict.RegisterObject(JsonExpObjectURI, &UrlValuesForJsonExp{Path: u.Path, UrlValues: u.Query()})
		return m.selectBackendJsonExp()
	default:
		return m.selectBackendMinPending()
	}
}

func (m *LblHttpClient) DoRequest(clientIp string, request *http.Request) (*http.Response, error) {
	backend, err := m.selectBackend(clientIp, request)
	if err != nil {
		return nil, err
	}
	request.URL.Host = backend.addr
	request.Host = backend.addr
	atomic.AddInt64(&backend.pendingRequests, 1)
	ret, err := backend.httpClient.Do(request)
	atomic.AddInt64(&backend.pendingRequests, -1)
	return ret, err
}
