// Copyright 2021 fangyousong(方友松). All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package lblhttpclient

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestLblHttpClient(t *testing.T) {
	lblC := NewLoadBalanceClient(MethodMinPending, 10, "hashkey", 0, 0)
	lblC.AddBackend("127.0.0.1:80", "default1", nil)
	lblC.AddBackend("127.0.0.1:80", "default2", nil)
	lblC.AddBackend("127.0.0.1:80", "default3", nil)

	req, _ := http.NewRequest("GET", "http://localhost/status", nil)
	resp, err := lblC.DoRequest("", req)
	if err != nil {
		t.Fatalf(err.Error())
	}
	bts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf(err.Error())
	}
	fmt.Println(string(bts))
}
