// Copyright 2021 fangyousong(方友松). All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package jsonexp

import (
	"crypto/md5"
	"fmt"
	"hash/fnv"
	"io"
	"strings"
)

const (
	PipelineFnLen      = "len"
	PipelineFnUpper    = "upper"
	PipelineFnLower    = "lower"
	PipelineFnFnv32    = "fnv32"
	PipelineFnFnv64    = "fnv64"
	PipelineFnMd5Lower = "md5"
	PipelineFnMd5Upper = "MD5"
)

func pipeFnLen(input interface{}, context Context) (interface{}, error) {
	if s, ok := GetStringValue(input); ok {
		return len(s), nil
	} else {
		return nil, fmt.Errorf("no string value")
	}
}

func pipeFnUpper(input interface{}, context Context) (interface{}, error) {
	if s, ok := GetStringValue(input); ok {
		return strings.ToUpper(s), nil
	} else {
		return nil, fmt.Errorf("no string value")
	}
}

func pipeFnLower(input interface{}, context Context) (interface{}, error) {
	if s, ok := GetStringValue(input); ok {
		return strings.ToLower(s), nil
	} else {
		return nil, fmt.Errorf("no string value")
	}
}

func pipeFnFnv32(input interface{}, context Context) (interface{}, error) {
	if s, ok := GetStringValue(input); ok {
		o := fnv.New32()
		io.WriteString(o, s)
		return o.Sum32(), nil
	} else {
		return nil, fmt.Errorf("no string value")
	}
}

func pipeFnFnv64(input interface{}, context Context) (interface{}, error) {
	if s, ok := GetStringValue(input); ok {
		o := fnv.New64()
		io.WriteString(o, s)
		return o.Sum64(), nil
	} else {
		return nil, fmt.Errorf("no string value")
	}
}

func pipeFnFnvMd5Lower(input interface{}, context Context) (interface{}, error) {
	if s, ok := GetStringValue(input); ok {
		h := md5.New()
		io.WriteString(h, s)
		return fmt.Sprintf("%x", h.Sum(nil)), nil
	} else {
		return nil, fmt.Errorf("no string value")
	}
}

func pipeFnFnvMd5Upper(input interface{}, context Context) (interface{}, error) {
	if s, ok := GetStringValue(input); ok {
		h := md5.New()
		io.WriteString(h, s)
		return fmt.Sprintf("%X", h.Sum(nil)), nil
	} else {
		return nil, fmt.Errorf("no string value")
	}
}

type PipeFunction func(input interface{}, context Context) (output interface{}, err error)

type pipeline struct {
	OriginName   string
	FunctionList []PipeFunction
}

func hasPipeline(varName string) bool {
	return strings.Contains(varName, "|")
}

func newPipeline(varName string, dict *Dictionary) (*pipeline, error) {
	if len(varName) < 2 || varName[:1] != "$" {
		return nil, fmt.Errorf("invalid var: %s", varName)
	}
	list := strings.Split(varName, "|")
	ret := &pipeline{}
	for i, v := range list {
		if i == 0 {
			ret.OriginName = v
		} else {
			fn := dict.GetPipeFunction(v)
			if fn != nil {
				ret.FunctionList = append(ret.FunctionList, fn)
			} else {
				return nil, fmt.Errorf("pip function %s not found", v)
			}
		}
	}
	return ret, nil
}

func (m *pipeline) Execute(originValue interface{}, context Context) (interface{}, error) {
	ret := originValue
	var err error
	for _, fn := range m.FunctionList {
		ret, err = fn(ret, context)
		if err != nil {
			return nil, err
		}
	}
	return ret, nil
}
