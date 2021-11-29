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
	ErrorAliasNotExist   = errors.New("backend's alias not found")
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

type HealthCheck func(req *http.Request, resp *http.Response, err error) bool

func DefaultHealthCheck(req *http.Request, resp *http.Response, err error) bool {
	return err == nil
}

func doInit() {
	var once sync.Once
	once.Do(func() {
		jsonExpDict = jsonexp.NewDictionary()
		jsonExpDict.RegisterVar(JsonExpVarTargetServer, nil)
	})
}

type lblHttpBackend struct {
	pendingRequests      int64
	healthCheck          HealthCheck
	healthCheckFailCount int64
	httpClient           *http.Client
	addr                 string // ip:port
	alias                string
}

func (m *lblHttpBackend) PendingRequests() int64 {
	return atomic.LoadInt64(&m.pendingRequests) + atomic.LoadInt64(&m.healthCheckFailCount)
}

func newLblHttpBackend(addr string, alias string, maxIdleConns int, connTimeout, waitTimeout time.Duration, healthCheck HealthCheck) (*lblHttpBackend, error) {
	if _, err := net.ResolveTCPAddr("tcp4", addr); err != nil {
		return nil, err
	}
	if healthCheck == nil {
		healthCheck = DefaultHealthCheck
	}
	ret := &lblHttpBackend{
		healthCheck: healthCheck,
		addr:        addr,
		alias:       alias,
		httpClient: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   connTimeout,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				ForceAttemptHTTP2: true,
				// MaxIdleConns:          maxIdleConns,
				MaxIdleConnsPerHost:   maxIdleConns,
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

func (m *LblHttpClient) AddBackend(addr string, alias string, healthCheck HealthCheck) error {
	if _, ok := m.serverMap[alias]; ok {
		return ErrorAliasExists
	}
	backend, err := newLblHttpBackend(addr, alias, m.maxIdleConnectionsPerServer, m.connectTimeout, m.waitResponseTimeout, healthCheck)
	if err != nil {
		return err
	}

	m.serverListLock.Lock()
	defer m.serverListLock.Unlock()
	m.serverList = append(m.serverList, backend)
	m.serverMap[alias] = backend

	return nil
}

func (m *LblHttpClient) RemoveBackend(alias string) error {
	m.serverListLock.Lock()
	defer m.serverListLock.Unlock()
	if _, ok := m.serverMap[alias]; ok {
		delete(m.serverMap, alias)
		for i, v := range m.serverList {
			if v.alias == alias {
				newList := m.serverList[:i]
				if i != len(m.serverList)-1 {
					newList = append(newList, m.serverList[i+1:]...)
				}
				m.serverList = newList
			}
		}
		return nil
	} else {
		return ErrorAliasNotExist
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
	if !backend.healthCheck(request, ret, err) {
		atomic.AddInt64(&backend.healthCheckFailCount, 1)
	} else {
		atomic.StoreInt64(&backend.healthCheckFailCount, backend.healthCheckFailCount/2)
	}

	return ret, err
}

func (m *LblHttpClient) getJsonExp() *jsonexp.JsonExpGroup {
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

func (m *LblHttpClient) selectBackendRoundrobin() (*lblHttpBackend, error) {
	idx := atomic.LoadInt64(&m.roundrobinIndex)
	atomic.AddInt64(&m.roundrobinIndex, 1)
	idx %= int64(len(m.serverList))
	ret := m.serverList[idx]
	return ret, nil
}

func (m *LblHttpClient) selectBackendRandom() (*lblHttpBackend, error) {
	idx := m.randObj.Intn(len(m.serverList))
	ret := m.serverList[idx]
	return ret, nil
}

func (m *LblHttpClient) selectBackendMinPending() (*lblHttpBackend, error) {
	idx := atomic.LoadInt64(&m.roundrobinIndex)
	atomic.AddInt64(&m.roundrobinIndex, 1)
	idx %= int64(len(m.serverList))
	minIdx := idx
	minPending := m.serverList[minIdx].PendingRequests()
	if minPending > 0 {
		for i := 0; i < len(m.serverList); i++ {
			idx++
			if idx >= int64(len(m.serverList)) {
				idx = 0
			}
			pr := m.serverList[idx].PendingRequests()
			if pr < minPending {
				minPending = pr
				minIdx = idx
			}
		}
	}
	return m.serverList[minIdx], nil
}

func (m *LblHttpClient) selectBackendIpHash(clientIp string) (*lblHttpBackend, error) {
	fnv32 := fnv.New32()
	io.WriteString(fnv32, clientIp)
	idx := fnv32.Sum32()
	idx = idx % uint32(len(m.serverList))
	return m.serverList[int(idx)], nil
}

func (m *LblHttpClient) selectBackendUrlParam(paramValue string) (*lblHttpBackend, error) {
	fnv32 := fnv.New32()
	io.WriteString(fnv32, paramValue)
	idx := fnv32.Sum32()
	idx = idx % uint32(len(m.serverList))
	return m.serverList[int(idx)], nil
}

func (m *LblHttpClient) selectBackendJsonExp(request *http.Request) (*lblHttpBackend, error) {
	jsonExp := m.getJsonExp()
	if jsonExp == nil {
		return nil, ErrorJsonExpNotFound
	}

	context := &jsonexp.DefaultContext{WithLock: false}
	u := request.URL
	jsonExpDict.RegisterObjectInContext(JsonExpObjectURI, &UrlValuesForJsonExp{Path: u.Path, UrlValues: u.Query()}, context)
	if err := jsonExp.Execute(context); err != nil {
		return nil, err
	}
	targetAlias, ok := context.GetCtxData(JsonExpVarTargetServer)
	if !ok {
		return nil, ErrorJsonExpNotFound
	}
	targetAliasStr, _ := jsonexp.GetStringValue(targetAlias)
	if backend, ok := m.serverMap[targetAliasStr]; ok {
		return backend, nil
	}
	return nil, ErrorJsonExpNotFound
}

func (m *LblHttpClient) selectBackend(clientIp string, request *http.Request) (*lblHttpBackend, error) {
	m.serverListLock.RLock()
	defer m.serverListLock.RUnlock()

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
		return m.selectBackendJsonExp(request)
	default:
		return m.selectBackendMinPending()
	}
}
