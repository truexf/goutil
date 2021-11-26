// Copyright 2021 fangyousong(方友松). All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package jsonexp

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/truexf/goutil"
)

type Context = goutil.Context

type DefaultContext = goutil.DefaultContext

type VarType uint

const (
	VarInvalid VarType = iota
	VarStr
	VarInt
	VarFloat
	VarSlice
)

func GetValueType(v interface{}) VarType {
	if v == nil {
		return VarInvalid
	}
	rv := reflect.ValueOf(v)
	switch kd := rv.Kind(); kd {
	case reflect.String:
		return VarStr
	case reflect.Float64, reflect.Float32:
		return VarFloat
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return VarInt
	case reflect.Slice:
		return VarSlice
	default:
		return VarInvalid
	}
}

func randomSelect(v []interface{}) interface{} {
	if len(v) == 0 {
		return nil
	}
	i := time.Now().Nanosecond() % len(v)
	return v[i]
}

func GetStringValue(v interface{}) (string, bool) {
	if v == nil {
		return "", true
	}
	rv := reflect.ValueOf(v)
	switch kd := rv.Kind(); kd {
	case reflect.Slice:
		r := randomSelect(v.([]interface{}))
		return GetStringValue(r)
	case reflect.String:
		return fmt.Sprintf("%s", v), true
	case reflect.Float64, reflect.Float32:
		return fmt.Sprintf("%f", v), true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return fmt.Sprintf("%d", v), true
	default:
		return "", false
	}
}

func GetFloatValue(v interface{}) (float64, bool) {
	if v == nil {
		return 0, true
	}
	rv := reflect.ValueOf(v)
	switch kd := rv.Kind(); kd {
	case reflect.Slice:
		r := randomSelect(v.([]interface{}))
		return GetFloatValue(r)
	case reflect.String:
		vs := v.(string)
		if ret, err := strconv.ParseFloat(vs, 64); err != nil {
			return 0, false
		} else {
			return ret, true
		}
	case reflect.Float64:
		return v.(float64), true
	case reflect.Float32:
		return float64(v.(float32)), true
	case reflect.Int:
		return float64(v.(int)), true
	case reflect.Int8:
		return float64(v.(int8)), true
	case reflect.Int16:
		return float64(v.(int16)), true
	case reflect.Int32:
		return float64(v.(int32)), true
	case reflect.Int64:
		return float64(v.(int64)), true
	case reflect.Uint:
		return float64(v.(uint)), true
	case reflect.Uint8:
		return float64(v.(uint8)), true
	case reflect.Uint16:
		return float64(v.(uint16)), true
	case reflect.Uint32:
		return float64(v.(uint32)), true
	case reflect.Uint64:
		return float64(v.(uint64)), true
	case reflect.Uintptr:
		return float64(v.(uintptr)), true
	default:
		return 0, false
	}
}

func GetIntValue(v interface{}) (int64, bool) {
	if v == nil {
		return 0, true
	}
	rv := reflect.ValueOf(v)
	switch kd := rv.Kind(); kd {
	case reflect.Slice:
		r := randomSelect(v.([]interface{}))
		return GetIntValue(r)
	case reflect.String:
		vs := v.(string)
		if ret, err := strconv.ParseInt(vs, 0, 64); err != nil {
			if retFloat, err := strconv.ParseFloat(vs, 64); err != nil {
				return 0, false
			} else {
				return int64(retFloat), true
			}
		} else {
			return ret, true
		}
	case reflect.Float64:
		return int64(rv.Float()), true
	case reflect.Float32:
		return int64(v.(float32)), true
	case reflect.Int:
		return int64(v.(int)), true
	case reflect.Int8:
		return int64(v.(int8)), true
	case reflect.Int16:
		return int64(v.(int16)), true
	case reflect.Int32:
		return int64(v.(int32)), true
	case reflect.Int64:
		return rv.Int(), true
	case reflect.Uint:
		return int64(v.(uint)), true
	case reflect.Uint8:
		return int64(v.(uint8)), true
	case reflect.Uint16:
		return int64(v.(uint16)), true
	case reflect.Uint32:
		return int64(v.(uint32)), true
	case reflect.Uint64:
		return int64(v.(uint64)), true
	case reflect.Uintptr:
		return int64(v.(uintptr)), true
	default:
		return 0, false
	}
}

type VarFunc func(context Context) (interface{}, error)
type CompareFunc func(leftValue interface{}, rightValue interface{}, context Context) (bool, error)
type AssignFunc func(varName string, leftValue interface{}, rightValue interface{}, context Context) error
type Object interface {
	GetPropertyValue(property string, context Context) interface{}
	SetPropertyValue(property string, value interface{}, context Context)
}

type Dictionary struct {
	varList              map[string]VarFunc
	varListLock          sync.RWMutex
	objectList           map[string]Object
	objectListLock       sync.RWMutex
	assignList           map[string]AssignFunc
	assignListLock       sync.RWMutex
	compareList          map[string]CompareFunc
	compareListLock      sync.RWMutex
	pipeFunctionList     map[string]PipeFunction
	pipeFunctionListLock sync.RWMutex
	regExpMacro          *regexp.Regexp
}

func NewDictionary() *Dictionary {
	ret := &Dictionary{varList: make(map[string]VarFunc),
		objectList:       make(map[string]Object),
		assignList:       make(map[string]AssignFunc),
		compareList:      make(map[string]CompareFunc),
		pipeFunctionList: make(map[string]PipeFunction),
		regExpMacro:      regexp.MustCompile(`\{\{(\$.+?)\}\}`),
	}
	ret.registerSystemPipeFunction()
	ret.registerSysemVariants()
	ret.registerSystemCompares()
	ret.registerSystemAssign()
	return ret
}

func (m *Dictionary) RegisterPipeFunction(name string, fn PipeFunction) {
	if name == "" || fn == nil {
		return
	}
	m.pipeFunctionListLock.Lock()
	defer m.pipeFunctionListLock.Unlock()
	m.pipeFunctionList[name] = fn
}

func (m *Dictionary) GetPipeFunction(name string) PipeFunction {
	m.pipeFunctionListLock.RLock()
	defer m.pipeFunctionListLock.RUnlock()
	if ret, ok := m.pipeFunctionList[name]; ok {
		return ret
	}
	return nil
}

func (dict *Dictionary) registerSysemVariants() {
	dict.RegisterVar("$datetime", DateTime)
	dict.RegisterVar("$date", Date)
	dict.RegisterVar("$time", Time)
	dict.RegisterVar("$stime", ShortTime)
	dict.RegisterVar("$year", Year)
	dict.RegisterVar("$month", Month)
	dict.RegisterVar("$day", Day)
	dict.RegisterVar("$hour", Hour)
	dict.RegisterVar("$minute", Minute)
	dict.RegisterVar("$second", Second)

	dict.RegisterVar("$iyear", IYear)
	dict.RegisterVar("$imonth", IMonth)
	dict.RegisterVar("$iday", IDay)
	dict.RegisterVar("$ihour", IHour)
	dict.RegisterVar("$iminute", IMinute)
	dict.RegisterVar("$isecond", ISecond)

	dict.RegisterVar("$rand", Rand) //1-100
	dict.RegisterVar("$break", nil)
}

func (dict *Dictionary) registerSystemCompares() {
	dict.RegisterCompare(">", More)
	dict.RegisterCompare(">=", MoreEqual)
	dict.RegisterCompare("<", Less)
	dict.RegisterCompare("<=", LessEqual)
	dict.RegisterCompare("=", Equal)
	dict.RegisterCompare("<>", NotEqual)
	dict.RegisterCompare("!=", NotEqual)
	dict.RegisterCompare("between", Between)
	dict.RegisterCompare("^between", NotBetween)
	dict.RegisterCompare("in", In)
	dict.RegisterCompare("not in", NotIn)
	dict.RegisterCompare("has", Has)
	dict.RegisterCompare("any", Any)
	dict.RegisterCompare("none", None)
	dict.RegisterCompare("~", Contain)
	dict.RegisterCompare("^~", NotContain)
	dict.RegisterCompare("~*", HeadMatch)
	dict.RegisterCompare("^~*", NotHeadMatch)
	dict.RegisterCompare("*~", TailMatch)
	dict.RegisterCompare("^*~", NotTailMatch)
	dict.RegisterCompare("cv", Cover)
	dict.RegisterCompare("^cv", NotCover)
}

func (dict *Dictionary) registerSystemAssign() {
	dict.RegisterAssign("=", Assign)
	dict.RegisterAssign("+=", AddAssign)
	dict.RegisterAssign("-=", SubAssign)
	dict.RegisterAssign("*=", MulAssign)
	dict.RegisterAssign("/=", DivAssign)
	dict.RegisterAssign("%=", ModAssign)
}

func (dict *Dictionary) registerSystemPipeFunction() {
	dict.RegisterPipeFunction(PipelineFnFnv32, pipeFnFnv32)
	dict.RegisterPipeFunction(PipelineFnFnv64, pipeFnFnv64)
	dict.RegisterPipeFunction(PipelineFnLen, pipeFnLen)
	dict.RegisterPipeFunction(PipelineFnUpper, pipeFnUpper)
	dict.RegisterPipeFunction(PipelineFnLower, pipeFnLower)
	dict.RegisterPipeFunction(PipelineFnMd5Lower, pipeFnFnvMd5Lower)
	dict.RegisterPipeFunction(PipelineFnMd5Upper, pipeFnFnvMd5Upper)
}

// 注册变量，变量名必须以"$"开头，且不能与object重名
func (m *Dictionary) RegisterVar(varName string, fetchFunc VarFunc) error {
	if varName == "" {
		return fmt.Errorf("varName is empty")
	}
	if _, ok := m.getObject(varName); ok {
		return fmt.Errorf("object with the same name exists")
	}
	m.varListLock.Lock()
	defer m.varListLock.Unlock()
	m.varList[varName] = fetchFunc
	return nil
}

// 注册对象，对象名必须以"$"开头，且不能与变量重名
func (m *Dictionary) RegisterObject(objectName string, object Object) error {
	if objectName == "" {
		return fmt.Errorf("objectName is empty")
	}
	// if _, ok := m.getVarFunc(objectName); ok {
	// 	return fmt.Errorf("variant with the same name exists")
	// }
	m.objectListLock.Lock()
	defer m.objectListLock.Unlock()
	m.objectList[objectName] = object
	return nil
}

// 注册条件运算符
func (m *Dictionary) RegisterCompare(compareName string, compareFunc CompareFunc) {
	if compareName == "" || compareFunc == nil {
		return
	}
	m.compareListLock.Lock()
	defer m.compareListLock.Unlock()
	m.compareList[compareName] = compareFunc
}

// 注册赋值运算符
func (m *Dictionary) RegisterAssign(assignName string, assignFunc AssignFunc) {
	if assignName == "" || assignFunc == nil {
		return
	}
	m.assignListLock.Lock()
	defer m.assignListLock.Unlock()
	m.assignList[assignName] = assignFunc
}

func (m *Dictionary) getVarFunc(varName string) (VarFunc, bool) {
	m.varListLock.RLock()
	defer m.varListLock.RUnlock()
	ret, ok := m.varList[varName]
	return ret, ok
}

func (m *Dictionary) getObject(objName string) (Object, bool) {
	m.objectListLock.RLock()
	defer m.objectListLock.RUnlock()
	ret, ok := m.objectList[objName]
	return ret, ok
}

func (m *Dictionary) getCompareFunc(compareName string) (CompareFunc, bool) {
	m.compareListLock.RLock()
	defer m.compareListLock.RUnlock()
	ret, ok := m.compareList[compareName]
	return ret, ok
}

func (m *Dictionary) getAssignFunc(assignName string) (AssignFunc, bool) {
	m.assignListLock.RLock()
	defer m.assignListLock.RUnlock()
	ret, ok := m.assignList[assignName]
	return ret, ok
}

// 获取对象的值， varName的格式为: ObjectName.PropertyName
func (m *Dictionary) getObjectPropertyValue(varName string, context Context) (interface{}, error) {
	parts := strings.Split(varName, ".")
	if len(parts) != 2 {
		return nil, fmt.Errorf("not object.property")
	}
	if obj, ok := m.getObject(parts[0]); ok {
		return obj.GetPropertyValue(parts[1], context), nil
	} else {
		return nil, fmt.Errorf("not an object")
	}
}

func (m *Dictionary) replaceMacro(strValueWithMacro string, context Context) string {
	matchedList := m.regExpMacro.FindAllStringSubmatch(strValueWithMacro, -1)
	if len(matchedList) == 0 {
		return strValueWithMacro
	}
	ret := strValueWithMacro
	for _, v := range matchedList {
		if len(v) != 2 {
			continue
		}
		macroValue, err := m.getOriginVarValue(v[1], context)
		if err == nil {
			if macroValueStr, ok := GetStringValue(macroValue); ok {
				ret = strings.Replace(ret, v[0], macroValueStr, 1)
			}
		}
	}
	return ret
}

func (m *Dictionary) getOriginVarValue(varName string, context Context) (interface{}, error) {
	fn, varFound := m.getVarFunc(varName)
	if !varFound {
		if ret, err := m.getObjectPropertyValue(varName, context); err == nil {
			return ret, nil
		} else {
			return nil, fmt.Errorf("variable(or object property) %s not found", varName)
		}
	}

	var ret interface{}
	var err error
	// use value from context first
	if r, ok := context.GetCtxData(varName); ok {
		ret = r
	} else if fn != nil {
		ret, err = fn(context)
	}
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (m *Dictionary) GetVarValue(varName string, context Context) (interface{}, error) {
	if varName == "" {
		return nil, fmt.Errorf("varName is empty")
	}
	if len(varName) <= 1 || varName[0] != '$' {
		return nil, fmt.Errorf("variable name must start with $")
	}
	if hasPipeline(varName) {
		varPipeline, err := newPipeline(varName, m)
		if err != nil {
			return nil, err
		}
		originValue, err := m.getOriginVarValue(varPipeline.OriginName, context)
		if err != nil {
			return nil, err
		}
		return varPipeline.Execute(originValue, context)
	} else {
		return m.getOriginVarValue(varName, context)
	}

}

func (m *Dictionary) Compare(compareName string, left string, right interface{}, context Context) (bool, error) {
	if compareName == "" {
		return false, fmt.Errorf("compare name is empty")
	}
	fn, ok := m.getCompareFunc(compareName)
	if !ok {
		return false, fmt.Errorf("compare name %s not found", compareName)
	}
	var leftValue interface{} = left
	var rightValue interface{} = right
	if len(left) > 1 && left[0] == '$' {
		leftValue, _ = m.GetVarValue(left, context)
	}
	if rightStr, ok := right.(string); ok {
		if len(rightStr) > 1 && rightStr[0] == '$' {
			rightValue, _ = m.GetVarValue(rightStr, context)
		}
	}
	if rightValueStr, ok := rightValue.(string); ok {
		rightValue = m.replaceMacro(rightValueStr, context)
	}
	return fn(leftValue, rightValue, context)
}

func (m *Dictionary) objectPropertyAssign(left string, right interface{}, context Context) bool {
	parts := strings.Split(left, ".")
	if len(parts) != 2 {
		return false
	}
	var rightValue interface{} = right
	if rightStr, ok := right.(string); ok {
		if len(rightStr) > 1 && rightStr[0] == '$' {
			rightValue, _ = m.GetVarValue(rightStr, context)
		}
	}
	if rightValueStr, ok := rightValue.(string); ok {
		rightValue = m.replaceMacro(rightValueStr, context)
	}
	if obj, ok := m.getObject(parts[0]); ok {
		obj.SetPropertyValue(parts[1], rightValue, context)
		return true
	}
	return false
}

func (m *Dictionary) Assign(assignName string, left string, right interface{}, context Context) error {
	if assignName == "" {
		return fmt.Errorf("assign name is empty")
	}

	if assignName == "=" && m.objectPropertyAssign(left, right, context) {
		return nil
	}

	fn, ok := m.getAssignFunc(assignName)
	if !ok {
		return fmt.Errorf("assign name %s not found", assignName)
	}
	leftValue, _ := m.GetVarValue(left, context)
	var rightValue interface{} = right
	if rightStr, ok := right.(string); ok {
		if len(rightStr) > 1 && rightStr[0] == '$' {
			rightValue, _ = m.GetVarValue(rightStr, context)
		}
	}
	if rightValueStr, ok := rightValue.(string); ok {
		rightValue = m.replaceMacro(rightValueStr, context)
	}
	return fn(left, leftValue, rightValue, context)
}

func (m *Dictionary) ListVars() []string {
	m.varListLock.RLock()
	defer m.varListLock.RUnlock()
	var ret []string
	for k := range m.varList {
		ret = append(ret, k)
	}
	return ret
}

func (m *Dictionary) ListObjects() []string {
	m.objectListLock.RLock()
	defer m.objectListLock.RUnlock()
	var ret []string
	for k := range m.objectList {
		ret = append(ret, k)
	}
	return ret
}

func (m *Dictionary) ListCompares() []string {
	m.compareListLock.RLock()
	defer m.compareListLock.RUnlock()
	var ret []string
	for k := range m.compareList {
		ret = append(ret, k)
	}
	return ret
}

func (m *Dictionary) ListAssigns() []string {
	m.assignListLock.RLock()
	defer m.assignListLock.RUnlock()
	var ret []string
	for k := range m.assignList {
		ret = append(ret, k)
	}
	return ret
}

type NameValue struct {
	Name  string
	Value interface{}
}

type AssignExp struct {
	Left       string
	Right      interface{}
	AssignName string
}

type CompareExp struct {
	Left        string
	Right       interface{}
	CompareName string
}

type JsonExp struct {
	compareExpList []*CompareExp
	assignExpList  []*AssignExp
	dict           *Dictionary
}

func (m *JsonExp) Execute(context Context) error {
	for _, v := range m.compareExpList {
		if ret, err := m.dict.Compare(v.CompareName, v.Left, v.Right, context); err != nil {
			return nil
		} else if !ret {
			return nil
		}
	}
	for _, v := range m.assignExpList {
		if err := m.dict.Assign(v.AssignName, v.Left, v.Right, context); err != nil {
			return err
		} else {
			if breaked, ok := context.GetCtxData("$break"); ok {
				if r, _ := GetIntValue(breaked); r == 1 {
					break
				}
			}
		}
	}
	return nil
}

func (m *JsonExp) GetCompareExpList() []*CompareExp {
	return m.compareExpList
}

func (m *JsonExp) GetAssignExpList() []*AssignExp {
	return m.assignExpList
}

/*
    a group of JsonExp
	json-exp group configuration json demo:
	{
		//json-exp group, from first to last, executing each json-exp node, breaking execution if variant $break = true, $break is a system variant
		"filter": [
			//json-exp node1, if all compare-exp return true, then execute assign-exp
			[
				["left-var","compare-name","right-var"], //compare-exp
				["left-var","compare-name","right-var"], //compare-exp
				["left-var","assign-name","right-var"], //the last exp is always an assign-exp
			],

			//json-exp node2
			[
				["left-var","compare-name","right-var"], //compare-exp
				["left-var","compare-name","right-var"], //compare-exp
				//multi-assign-exp
				[
					["left-var","assign-name","right-var"],
					["left-var","assign-name","right-var"],
					["left-var","assign-name","right-var"]
				]
			],

			//json-exp node3
			[
				//...
			]
		]
	}
*/
type JsonExpGroup struct {
	dict        *Dictionary
	groupSource interface{}
	group       []*JsonExp
}

func NewJsonExpGroup(dict *Dictionary, groupSource interface{}) (*JsonExpGroup, error) {
	if dict == nil || groupSource == nil {
		return nil, fmt.Errorf("nil dict or groupSource")
	}
	ret := &JsonExpGroup{dict: dict, groupSource: groupSource}
	if err := ret.parse(); err != nil {
		return nil, err
	}
	return ret, nil
}

func (m *JsonExpGroup) parse() error {
	v := reflect.ValueOf(m.groupSource)
	if v.Kind() != reflect.Slice {
		return fmt.Errorf("invalid groupSource, not a slice")
	}
	group := m.groupSource.([]interface{})
	for _, nodeSource := range group {
		v := reflect.ValueOf(nodeSource)
		if v.Kind() != reflect.Slice {
			return fmt.Errorf("invalid groupSource, exp node is not a slice")
		}
		node := nodeSource.([]interface{})
		jsonExp := &JsonExp{dict: m.dict}
		for i, expSource := range node {
			v := reflect.ValueOf(expSource)
			if v.Kind() != reflect.Slice {
				return fmt.Errorf("invalid groupSource, exp is not a slice")
			}
			exp := expSource.([]interface{})
			if i == len(node)-1 {
				//assign exp
				isMultiAssign := true
				for _, assignExpSource := range exp {
					if !(reflect.ValueOf(assignExpSource).Kind() == reflect.Slice && len(assignExpSource.([]interface{})) == 3) {
						isMultiAssign = false
						break
					}
				}
				if !isMultiAssign {
					if len(exp) != 3 {
						return fmt.Errorf("invalid groupSource, len(exp) <> 3")
					}
					if reflect.ValueOf(exp[0]).Kind() != reflect.String {
						return fmt.Errorf("invalid groupSource, exp[0] is not string")
					}
					varName := exp[0].(string)
					if len(varName) <= 1 || varName[0] != '$' {
						return fmt.Errorf("invalid groupSource, exp[0] is not a variant")
					}
					if reflect.ValueOf(exp[1]).Kind() != reflect.String {
						return fmt.Errorf("invalid groupSource, exp[1] is not string")
					}
					assignExp := &AssignExp{Left: exp[0].(string), Right: exp[2], AssignName: exp[1].(string)}
					jsonExp.assignExpList = append(jsonExp.assignExpList, assignExp)
				} else {
					for _, internalExp := range exp {
						exp := internalExp.([]interface{})
						if len(exp) != 3 {
							return fmt.Errorf("invalid groupSource, len(exp) <> 3")
						}
						if reflect.ValueOf(exp[0]).Kind() != reflect.String {
							return fmt.Errorf("invalid groupSource, exp[0] is not string")
						}
						varName := exp[0].(string)
						if len(varName) <= 1 || varName[0] != '$' {
							return fmt.Errorf("invalid groupSource, exp[0] is not a variant")
						}
						if reflect.ValueOf(exp[1]).Kind() != reflect.String {
							return fmt.Errorf("invalid groupSource, exp[1] is not string")
						}
						assignExp := &AssignExp{Left: exp[0].(string), Right: exp[2], AssignName: exp[1].(string)}
						jsonExp.assignExpList = append(jsonExp.assignExpList, assignExp)
					}
				}
			} else {
				//compare exp
				if len(exp) != 3 {
					return fmt.Errorf("invalid groupSource, len(exp) <> 3")
				}
				if reflect.ValueOf(exp[0]).Kind() != reflect.String {
					return fmt.Errorf("invalid groupSource, exp[0] is not string")
				}
				varName := exp[0].(string)
				if len(varName) <= 1 || varName[0] != '$' {
					return fmt.Errorf("invalid groupSource, exp[0] is not a variant")
				}
				if reflect.ValueOf(exp[1]).Kind() != reflect.String {
					return fmt.Errorf("invalid groupSource, exp[1] is not string")
				}
				compareExp := &CompareExp{Left: exp[0].(string), Right: exp[2], CompareName: exp[1].(string)}
				jsonExp.compareExpList = append(jsonExp.compareExpList, compareExp)
			}
		}
		m.group = append(m.group, jsonExp)
	}

	return nil
}

// 执行表达式组
func (m *JsonExpGroup) Execute(context Context) error {
	if context != nil {
		if _, ok := context.GetCtxData("$rand"); !ok {
			context.SetCtxData("$rand", rand.Intn(100)+1)
		}
	}
	for _, jsonExp := range m.group {
		if err := jsonExp.Execute(context); err != nil {
			return err
		} else {
			if breaked, ok := context.GetCtxData("$break"); ok {
				if r, _ := GetIntValue(breaked); r == 1 {
					break
				}
			}
		}
	}
	return nil
}

func (m *JsonExpGroup) List() []*JsonExp {
	return m.group
}

// Configuration对象代表一个json配置,其中包含，0个或n个key/value键值对，以及0个或n个JSON表达式组
type Configuration struct {
	jsonSource    []byte
	dict          *Dictionary
	nameValues    map[string]interface{}
	jsonExpGroups map[string]*JsonExpGroup
}

// 传入json,创建一个Configuration对象
func NewConfiguration(jsonSource []byte, dict *Dictionary) (*Configuration, error) {
	if len(jsonSource) == 0 || dict == nil {
		return nil, fmt.Errorf("invalid jsonSource or dict")
	}
	mp := make(map[string]interface{})
	if err := json.Unmarshal(jsonSource, &mp); err != nil {
		return nil, fmt.Errorf("unmarshal json fail, %s", err.Error())
	}
	ret := &Configuration{
		jsonSource:    jsonSource,
		dict:          dict,
		nameValues:    make(map[string]interface{}),
		jsonExpGroups: make(map[string]*JsonExpGroup),
	}
	for k, v := range mp {
		if group, err := NewJsonExpGroup(dict, v); err == nil {
			ret.jsonExpGroups[k] = group
		} else {
			ret.nameValues[k] = v
		}
	}
	return ret, nil
}

// 获取键值
func (m *Configuration) GetNameValue(key string, context Context) (interface{}, bool) {
	ret, ok := m.nameValues[key]
	if !ok {
		return nil, false
	}
	if retStr, ok := ret.(string); ok {
		if len(retStr) > 1 && retStr[0] == '$' {
			if retValue, err := m.dict.GetVarValue(retStr, context); err == nil {
				return retValue, true
			}
		}
	}
	return ret, true
}

// 获取JSON表达式组
func (m *Configuration) GetJsonExpGroup(key string) (*JsonExpGroup, bool) {
	ret, ok := m.jsonExpGroups[key]
	if !ok {
		return nil, false
	}
	return ret, true
}
