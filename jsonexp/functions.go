// Copyright 2021 fangyousong(方友松). All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package jsonexp

import (
	"fmt"
	"github.com/truexf/goutil"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var DateTime = func(context Context) (interface{}, error) {
	return time.Now().Format("2006-01-02 15:04:05"), nil
}

var Date = func(context Context) (interface{}, error) {
	return time.Now().Format("2006-01-02"), nil
}

var Time = func(context Context) (interface{}, error) {
	return time.Now().Format("15:04:05"), nil
}

var ShortTime = func(context Context) (interface{}, error) {
	return time.Now().Format("15:04"), nil
}

var Year = func(context Context) (interface{}, error) {
	return time.Now().Format("2006"), nil
}

var IYear = func(context Context) (interface{}, error) {
	return time.Now().Year(), nil
}

var Month = func(context Context) (interface{}, error) {
	return time.Now().Format("01"), nil
}
var IMonth = func(context Context) (interface{}, error) {
	return time.Now().Month(), nil
}

var Day = func(context Context) (interface{}, error) {
	return time.Now().Format("02"), nil
}

var IDay = func(context Context) (interface{}, error) {
	return time.Now().Day(), nil
}

var Hour = func(context Context) (interface{}, error) {
	return time.Now().Format("15"), nil
}

var IHour = func(context Context) (interface{}, error) {
	return time.Now().Hour(), nil
}

var Minute = func(context Context) (interface{}, error) {
	return time.Now().Format("04"), nil
}

var IMinute = func(context Context) (interface{}, error) {
	return time.Now().Minute(), nil
}

var Second = func(context Context) (interface{}, error) {
	return time.Now().Format("05"), nil
}

var ISecond = func(context Context) (interface{}, error) {
	return time.Now().Second(), nil
}

var Rand = func(context Context) (interface{}, error) {
	if ret, ok := context.GetCtxData("$rand"); ok {
		if rd, ok := GetIntValue(ret); ok {
			return rd, nil
		}
	}
	return rand.Intn(100) + 1, nil
}

// compares
var NotCover = func(L, R interface{}, context Context) (bool, error) {
	ret, err := Cover(L, R, context)
	if err != nil {
		return false, err
	}
	return !ret, err
}

var Cover = func(L, R interface{}, context Context) (bool, error) {
	l, lOk := GetStringValue(L)
	if !lOk {
		return false, fmt.Errorf("invalid L")
	}
	r, rOk := GetStringValue(R)
	if !rOk {
		return false, fmt.Errorf("right value is not string incompatible")
	}

	if l == "" || r == "" {
		return false, nil
	}

	rList := strings.Split(r, ",")
	for _, v := range rList {
		if strings.Contains(l, v) {
			return true, nil
		}
	}
	return false, nil
}

var Contain = func(L, R interface{}, context Context) (bool, error) {
	l, lOk := GetStringValue(L)
	if !lOk {
		return false, fmt.Errorf("invalid L")
	}
	r, rOk := GetStringValue(R)
	if !rOk {
		return false, fmt.Errorf("right value not string-incompatible")
	}

	if l == "" || r == "" {
		return false, nil
	}

	ret := strings.Contains(l, r)
	return ret, nil
}

var RegExpMatch = func(L, R interface{}, context Context) (bool, error) {
	l, lOk := GetStringValue(L)
	if !lOk {
		return false, fmt.Errorf("invalid L")
	}
	r, rOk := GetStringValue(R)
	if !rOk {
		return false, fmt.Errorf("right value not string-incompatible")
	}

	if l == "" || r == "" {
		return false, nil
	}

	if matched, err := regexp.Match(r, goutil.UnsafeStringToBytes(l)); err == nil {
		return matched, nil
	} else {
		return false, err
	}
}

var NotContain = func(L, R interface{}, context Context) (bool, error) {
	ret, err := Contain(L, R, context)
	if err != nil {
		return false, err
	}
	return !ret, nil
}

var HeadMatch = func(L, R interface{}, context Context) (bool, error) {
	l, lOk := GetStringValue(L)
	if !lOk {
		return false, fmt.Errorf("invalid L")
	}
	r, rOk := GetStringValue(R)
	if !rOk {
		return false, fmt.Errorf("right value not string-incompatible")
	}

	if l == "" || r == "" {
		return false, nil
	}

	ret := strings.Index(l, r) == 0
	return ret, nil
}

var NotHeadMatch = func(L, R interface{}, context Context) (bool, error) {
	ret, err := HeadMatch(L, R, context)
	if err != nil {
		return false, err
	}
	return !ret, nil
}

var TailMatch = func(L, R interface{}, context Context) (bool, error) {
	l, lOk := GetStringValue(L)
	if !lOk {
		return false, fmt.Errorf("invalid L")
	}
	r, rOk := GetStringValue(R)
	if !rOk {
		return false, fmt.Errorf("right value not string-incompatible")
	}

	if l == "" || r == "" {
		return false, nil
	}

	ret := strings.Index(l, r) == len(l)-len(r)
	return ret, nil
}

var NotTailMatch = func(L, R interface{}, context Context) (bool, error) {
	ret, err := TailMatch(L, R, context)
	if err != nil {
		return false, err
	}
	return !ret, nil
}

var None = func(L, R interface{}, context Context) (bool, error) {
	any, err := Any(L, R, context)
	if err != nil {
		return false, err
	}
	return !any, nil
}

var Any = func(L, R interface{}, context Context) (bool, error) {
	l, lOk := GetStringValue(L)
	if !lOk {
		return false, fmt.Errorf("invalid L")
	}
	r, rOk := GetStringValue(R)
	if !rOk {
		return false, fmt.Errorf("right value not string-incompatible")
	}

	if l == "" || r == "" {
		return false, nil
	}

	lList := strings.Split(l, ",")
	rList := strings.Split(r, ",")
	any := false
	for _, v := range rList {
		for _, vL := range lList {
			if v == vL {
				any = true
				break
			}
		}
		if any {
			return true, nil
		}
	}
	return false, nil
}

var Has = func(L, R interface{}, context Context) (bool, error) {
	l, lOk := GetStringValue(L)
	if !lOk {
		return false, fmt.Errorf("invalid L")
	}
	r, rOk := GetStringValue(R)
	if !rOk {
		return false, fmt.Errorf("right value not string-incompatible")
	}

	if l == "" || r == "" {
		return false, nil
	}

	lList := strings.Split(l, ",")
	rList := strings.Split(r, ",")
	for _, v := range rList {
		has := false
		for _, vL := range lList {
			if v == vL {
				has = true
				break
			}
		}
		if !has {
			return false, nil
		}
	}
	return true, nil
}

var NotIn = func(L, R interface{}, context Context) (bool, error) {
	l, lOk := GetStringValue(L)
	if !lOk {
		return false, fmt.Errorf("invalid L")
	}
	if r, ok := GetStringValue(R); !ok {
		return false, fmt.Errorf("right value not string-incompatible")
	} else {
		rList := strings.Split(r, ",")
		in := false
		for _, v := range rList {
			if l == v {
				in = true
				break
			}
		}
		return !in, nil
	}
}

var In = func(L, R interface{}, context Context) (bool, error) {
	l, lOk := GetStringValue(L)
	if !lOk {
		return false, fmt.Errorf("invalid L")
	}
	if r, ok := GetStringValue(R); !ok {
		return false, fmt.Errorf("right value not string-incompatible")
	} else {
		rList := strings.Split(r, ",")
		for _, v := range rList {
			if l == v {
				return true, nil
			}
		}
	}
	return false, nil
}

var NotBetween = func(L, R interface{}, context Context) (bool, error) {
	ret, err := Between(L, R, context)
	if err != nil {
		return false, err
	}
	return !ret, nil
}

var Between = func(L, R interface{}, context Context) (bool, error) {
	tp := GetValueType(L)
	if tp == VarStr {
		l, _ := GetStringValue(L)
		if r, ok := GetStringValue(R); !ok {
			return false, fmt.Errorf("right value not string-incompatible")
		} else {
			rList := strings.Split(r, ",")
			if len(rList) != 2 {
				return false, fmt.Errorf("right value not a between string")
			} else {
				return l >= rList[0] && l <= rList[1], nil
			}
		}
	} else if tp == VarInt {
		l, _ := GetIntValue(L)
		if r, ok := GetStringValue(R); !ok {
			return false, fmt.Errorf("right value not a between string")
		} else {
			rList := strings.Split(r, ",")
			if len(rList) != 2 {
				return false, fmt.Errorf("right value not a between string")
			} else {
				if b, errB := strconv.ParseInt(rList[0], 0, 64); errB == nil {
					if e, errE := strconv.ParseInt(rList[1], 0, 64); errE == nil {
						return l >= b && l <= e, nil
					}
				}
				return false, fmt.Errorf("right value not a between string")
			}
		}
	} else if tp == VarFloat {
		l, _ := GetFloatValue(L)
		if r, ok := GetStringValue(R); !ok {
			return false, fmt.Errorf("right value not a between string")
		} else {
			rList := strings.Split(r, ",")
			if len(rList) != 2 {
				return false, fmt.Errorf("right value not a between string")
			} else {
				if b, errB := strconv.ParseFloat(rList[0], 64); errB == nil {
					if e, errE := strconv.ParseFloat(rList[1], 64); errE == nil {
						return l >= b && l <= e, nil
					}
				}
				return false, fmt.Errorf("right value not a between string")
			}
		}
	} else {
		return false, fmt.Errorf("invalid data type of L")
	}
}

var NotEqual = func(L, R interface{}, context Context) (bool, error) {
	tp := GetValueType(L)
	if tp == VarInvalid {
		return false, fmt.Errorf("invalid param L")
	} else if tp == VarStr {
		s, _ := GetStringValue(L)
		if sR, ok := GetStringValue(R); ok {
			return s != sR, nil
		} else {
			return false, fmt.Errorf("right value not string-incompatible")
		}
	} else if tp == VarInt {
		i, _ := GetIntValue(L)
		if iR, ok := GetIntValue(R); ok {
			return i != iR, nil
		} else {
			return false, fmt.Errorf("right value not int-incompatible")
		}
	} else if tp == VarFloat {
		f, _ := GetFloatValue(L)
		if fR, ok := GetFloatValue(R); ok {
			return f != fR, nil
		} else {
			return false, fmt.Errorf("right value not float-incompatible")
		}
	}
	return false, nil
}

var Equal = func(L, R interface{}, context Context) (bool, error) {
	tp := GetValueType(L)
	if tp == VarInvalid {
		return false, fmt.Errorf("invalid param L")
	} else if tp == VarStr {
		s, _ := GetStringValue(L)
		if sR, ok := GetStringValue(R); ok {
			return s == sR, nil
		} else {
			return false, fmt.Errorf("right value not string-incompatible")
		}
	} else if tp == VarInt {
		i, _ := GetIntValue(L)
		if iR, ok := GetIntValue(R); ok {
			return i == iR, nil
		} else {
			return false, fmt.Errorf("right value not int-incompatible")
		}
	} else if tp == VarFloat {
		f, _ := GetFloatValue(L)
		if fR, ok := GetFloatValue(R); ok {
			return f == fR, nil
		} else {
			return false, fmt.Errorf("right value not float-incompatible")
		}
	}
	return false, nil
}

var LessEqual = func(L, R interface{}, context Context) (bool, error) {
	tp := GetValueType(L)
	if tp == VarInvalid {
		return false, fmt.Errorf("invalid param L")
	} else if tp == VarStr {
		s, _ := GetStringValue(L)
		if sR, ok := GetStringValue(R); ok {
			return s <= sR, nil
		} else {
			return false, fmt.Errorf("right value not string-incompatible")
		}
	} else if tp == VarInt {
		i, _ := GetIntValue(L)
		if iR, ok := GetIntValue(R); ok {
			return i <= iR, nil
		} else {
			return false, fmt.Errorf("right value not int-incompatible")
		}
	} else if tp == VarFloat {
		f, _ := GetFloatValue(L)
		if fR, ok := GetFloatValue(R); ok {
			return f <= fR, nil
		} else {
			return false, fmt.Errorf("right value not float-incompatible")
		}
	}
	return false, nil
}

var Less = func(L, R interface{}, context Context) (bool, error) {
	tp := GetValueType(L)
	if tp == VarInvalid {
		return false, fmt.Errorf("invalid param L")
	} else if tp == VarStr {
		s, _ := GetStringValue(L)
		if sR, ok := GetStringValue(R); ok {
			return s < sR, nil
		} else {
			return false, fmt.Errorf("right value not string-incompatible")
		}
	} else if tp == VarInt {
		i, _ := GetIntValue(L)
		if iR, ok := GetIntValue(R); ok {
			return i < iR, nil
		} else {
			return false, fmt.Errorf("right value not int-incompatible")
		}
	} else if tp == VarFloat {
		f, _ := GetFloatValue(L)
		if fR, ok := GetFloatValue(R); ok {
			return f < fR, nil
		} else {
			return false, fmt.Errorf("right value not float-incompatible")
		}
	}
	return false, nil
}

var MoreEqual = func(L, R interface{}, context Context) (bool, error) {
	tp := GetValueType(L)
	if tp == VarInvalid {
		return false, fmt.Errorf("invalid param L")
	} else if tp == VarStr {
		s, _ := GetStringValue(L)
		if sR, ok := GetStringValue(R); ok {
			return s >= sR, nil
		} else {
			return false, fmt.Errorf("right value not string-incompatible")
		}
	} else if tp == VarInt {
		i, _ := GetIntValue(L)
		if iR, ok := GetIntValue(R); ok {
			return i >= iR, nil
		} else {
			return false, fmt.Errorf("right value not int-incompatible")
		}
	} else if tp == VarFloat {
		f, _ := GetFloatValue(L)
		if fR, ok := GetFloatValue(R); ok {
			return f >= fR, nil
		} else {
			return false, fmt.Errorf("right value not float-incompatible")
		}
	}
	return false, nil
}

var More = func(L, R interface{}, context Context) (bool, error) {
	tp := GetValueType(L)
	if tp == VarInvalid {
		return false, fmt.Errorf("invalid param L")
	} else if tp == VarStr {
		s, _ := GetStringValue(L)
		if sR, ok := GetStringValue(R); ok {
			return s > sR, nil
		} else {
			return false, fmt.Errorf("right value not string-incompatible")
		}
	} else if tp == VarInt {
		i, _ := GetIntValue(L)
		if iR, ok := GetIntValue(R); ok {
			return i > iR, nil
		} else {
			return false, fmt.Errorf("right value not int-incompatible")
		}
	} else if tp == VarFloat {
		f, _ := GetFloatValue(L)
		if fR, ok := GetFloatValue(R); ok {
			return f > fR, nil
		} else {
			return false, fmt.Errorf("right value not float-incompatible")
		}
	}
	return false, nil
}

// assign functions
var Assign = func(L string, lValue interface{}, R interface{}, ret Context) error {
	if ret == nil {
		return fmt.Errorf("param ret is nil")
	}
	ret.SetCtxData(L, R)
	return nil
}

var AddAssign = func(L string, lValue interface{}, R interface{}, ret Context) error {
	if ret == nil {
		return fmt.Errorf("param ret is nil")
	}

	lType := GetValueType(lValue)
	vType := GetValueType(R)
	if lType == VarInvalid {
		lType = vType
	}
	switch lType {
	case VarStr:
		old, _ := GetStringValue(lValue)
		add, _ := GetStringValue(R)
		ret.SetCtxData(L, old+add)
	case VarFloat:
		old, _ := GetFloatValue(lValue)
		add, _ := GetFloatValue(R)
		ret.SetCtxData(L, old+add)
	case VarInt:
		old, _ := GetIntValue(lValue)
		add, _ := GetIntValue(R)
		ret.SetCtxData(L, old+add)
	default:
		return fmt.Errorf("invalid operand")
	}
	return nil
}

var SubAssign = func(L string, lValue interface{}, R interface{}, ret Context) error {
	if ret == nil {
		return fmt.Errorf("param ret is nil")
	}

	vType := GetValueType(R)
	switch vType {
	case VarFloat:
		old, _ := GetFloatValue(lValue)
		ret.SetCtxData(L, old-R.(float64))
	case VarInt:
		old, _ := GetIntValue(lValue)
		ret.SetCtxData(L, old-R.(int64))
	default:
		return fmt.Errorf("invalid operand")
	}
	return nil
}

var MulAssign = func(L string, lValue interface{}, R interface{}, ret Context) error {
	if ret == nil {
		return fmt.Errorf("param ret is nil")
	}

	lType := GetValueType(lValue)
	vType := GetValueType(R)
	if lType == VarInvalid {
		lType = vType
	}
	switch lType {
	case VarStr:
		if vType == VarInt {
			old, _ := GetStringValue(lValue)
			ret.SetCtxData(L, strings.Repeat(old, R.(int)))
		} else {
			return fmt.Errorf("invalid operand")
		}
	case VarFloat:
		old, _ := GetFloatValue(lValue)
		ret.SetCtxData(L, old*R.(float64))
	case VarInt:
		old, _ := GetIntValue(lValue)
		ret.SetCtxData(L, old*R.(int64))
	default:
		return fmt.Errorf("invalid operand")
	}
	return nil
}

var DivAssign = func(L string, lValue interface{}, R interface{}, ret Context) error {
	if ret == nil {
		return fmt.Errorf("param ret is nil")
	}

	old, _ := GetFloatValue(lValue)
	add, _ := GetFloatValue(R)
	if add <= 0.00001 && add >= -0.00001 {
		return fmt.Errorf("invalid operand")
	}
	ret.SetCtxData(L, old/add)
	return nil
}

var ModAssign = func(L string, lValue interface{}, R interface{}, ret Context) error {
	if ret == nil {
		return fmt.Errorf("param ret is nil")
	}

	old, _ := GetFloatValue(lValue)
	add, _ := GetFloatValue(R)
	addInt := int64(add)
	if addInt == 0 {
		return fmt.Errorf("invalid operand")
	}
	ret.SetCtxData(L, int64(old)%addInt)
	return nil
}
