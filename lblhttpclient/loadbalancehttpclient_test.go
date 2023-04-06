// Copyright 2021 fangyousong(方友松). All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package lblhttpclient

import (
	"fmt"
	"github.com/truexf/goutil"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestLblHttpClientMinPending(t *testing.T) {
	lblC := NewLoadBalanceClient(MethodMinPending, 10, "hashkey", 0, 0)
	lblC.AddBackend("127.0.0.1:81", "default1", nil)
	lblC.AddBackend("127.0.0.1:80", "default2", nil)
	lblC.AddBackend("127.0.0.1:80", "default3", nil)

	req, _ := http.NewRequest("GET", "http://localhost/status", nil)
	for i := 0; i < 10; i++ {
		doRequest("minpending", "127.0.0.1", req, lblC)
	}
}

func TestLblHttpClientRoundrobin(t *testing.T) {
	lblC := NewLoadBalanceClient(MethodRoundrobin, 10, "hashkey", 0, 0)
	lblC.AddBackend("127.0.0.1:81", "default1", nil)
	lblC.AddBackend("127.0.0.1:80", "default2", nil)
	lblC.AddBackend("127.0.0.1:80", "default3", nil)

	req, _ := http.NewRequest("GET", "http://localhost/status", nil)
	for i := 0; i < 10; i++ {
		doRequest("roundrobin", "127.0.0.1", req, lblC)
	}
}

func TestLblHttpClientRandom(t *testing.T) {
	lblC := NewLoadBalanceClient(MethodRandom, 10, "hashkey", 0, 0)
	lblC.AddBackend("127.0.0.1:81", "default1", nil)
	lblC.AddBackend("127.0.0.1:80", "default2", nil)
	lblC.AddBackend("127.0.0.1:80", "default3", nil)

	req, _ := http.NewRequest("GET", "http://localhost/status", nil)
	for i := 0; i < 10; i++ {
		doRequest("random", "127.0.0.1", req, lblC)
	}
}

func TestLblHttpClientIpHash(t *testing.T) {
	lblC := NewLoadBalanceClient(MethodIpHash, 10, "hashkey", 0, 0)
	lblC.AddBackend("127.0.0.1:81", "default1", nil)
	lblC.AddBackend("127.0.0.1:80", "default2", nil)
	lblC.AddBackend("127.0.0.1:80", "default3", nil)

	req, _ := http.NewRequest("GET", "http://localhost/status", nil)
	for i := 0; i < 10; i++ {
		doRequest("iphash", "127.0.0.1", req, lblC)
	}
}

func TestLblHttpClientUrlparam(t *testing.T) {
	lblC := NewLoadBalanceClient(MethodUrlParam, 10, "hashkey", 0, 0)
	lblC.AddBackend("127.0.0.1:81", "default1", nil)
	lblC.AddBackend("127.0.0.1:80", "default2", nil)
	lblC.AddBackend("127.0.0.1:80", "default3", nil)

	req, _ := http.NewRequest("GET", "http://localhost/status?hashkey=value", nil)
	for i := 0; i < 10; i++ {
		doRequest("urlparam", "127.0.0.1", req, lblC)
	}
}

func TestLblHttpClientJsonExp(t *testing.T) {
	lblC := NewLoadBalanceClient(MethodJsonExp, 10, "hashkey", 0, 0)
	lblC.SetJsonExp(goutil.UnsafeStringToBytes(jsExpString))
	lblC.AddBackend("127.0.0.1:81", "default1", nil)
	lblC.AddBackend("127.0.0.1:80", "default2", nil)
	lblC.AddBackend("127.0.0.1:80", "default3", nil)

	req, _ := http.NewRequest("GET", "http://localhost/status?hashkey=value", nil)
	doRequest("jsonexp", "127.0.0.1", req, lblC)
	req, _ = http.NewRequest("GET", "http://localhost/not_status?hashkey=value", nil)
	doRequest("jsonexp", "127.0.0.1", req, lblC)
}

func BenchmarkJsonExp(b *testing.B) {
	lblC := NewLoadBalanceClient(MethodJsonExp, 10, "hashkey", 0, 0)
	lblC.SetJsonExp(goutil.UnsafeStringToBytes(jsExpString))
	lblC.AddBackend("127.0.0.1:81", "default1", nil)
	lblC.AddBackend("127.0.0.1:80", "default2", nil)
	lblC.AddBackend("127.0.0.1:80", "default3", nil)

	req, _ := http.NewRequest("GET", "http://localhost/not_status?hashkey=value", nil)
	benchmark = true
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doRequest("jsonexp", "127.0.0.1", req, lblC)
		}
	})

}

func doRequest(tp string, ip string, req *http.Request, lblC *LblHttpClient) {
	resp, err := lblC.DoRequest(ip, req)
	if benchmark {
		return
	}
	if err != nil {
		fmt.Printf("%s failed: %s\n", tp, err.Error())
		return
	}
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("%s success.\n", tp)
}

var benchmark bool
var jsExpString = `{
	"LB_TARGET":[
		[
			["$REQUEST_URI.__PATH__", "=", "/status"],
			[
				["$LB_TARGET_SERVER", "=", "default1"],
				["$break", "=", 1]
			]
		],
		[
			["$REQUEST_URI.__PATH__", "<>", "/status"],
			["$LB_TARGET_SERVER", "=", "default2"]
		]
	]	
}`
