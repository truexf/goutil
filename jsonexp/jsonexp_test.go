package jsonexp

import (
	"fmt"
	"testing"
	"time"
)

type MyObj struct {
	ver string
}

func (m *MyObj) GetPropertyValue(PropertyName string, context Context) interface{} {
	return m.ver
}

func (m *MyObj) SetPropertyValue(property string, value interface{}, context Context) {
	if s, ok := GetStringValue(value); ok {
		m.ver = s
	}
}

func TestJsonExp(t *testing.T) {
	dict := NewDictionary()
	dict.RegisterVar("$my_var", nil)
	myobj := &MyObj{}
	dict.RegisterObject("$myobj", myobj)
	cfg, err := NewConfiguration([]byte(jsonSource), dict)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if v, _ := cfg.GetNameValue("name1", nil); v.(string) != "value1" {
		t.Fatalf("name value fail")
	}
	if v, _ := cfg.GetNameValue("name2", nil); int(v.(float64)) != 1234 {
		t.Fatalf("name value fail")
	}
	if v, _ := cfg.GetNameValue("name3", nil); v.(bool) != true {
		t.Fatalf("name value fail")
	}
	if v, _ := cfg.GetNameValue("name4", nil); len(v.([]interface{})) != 2 {
		t.Fatalf("name value fail")
	}
	// fmt.Printf("%v\n", cfg.)
	g, ok := cfg.GetJsonExpGroup("my_json_exp_group")
	if !ok {
		t.Fatalf("no jsonexp group")
	}
	ctx := &DefaultContext{}
	ctx.SetCtxData("$rand", time.Now().Second()%10)
	g.Execute(ctx)
	myVar, _ := ctx.GetCtxData("$my_var")
	fmt.Printf("%v\n", myVar)
	fmt.Printf("myobj.ver: %s\n", myobj.ver)
}

func BenchmarkJsonExp(b *testing.B) {
	dict := NewDictionary()
	dict.RegisterVar("$my_var", nil)
	cfg, err := NewConfiguration([]byte(jsonSource), dict)
	if err != nil {
		b.Fatalf(err.Error())
	}
	// fmt.Printf("%v\n", cfg.)
	g, ok := cfg.GetJsonExpGroup("my_json_exp_group")
	if !ok {
		b.Fatalf("no jsonexp group")
	}
	ctx := &DefaultContext{}
	ctx.SetCtxData("$rand", time.Now().Second()%10)
	for i := 0; i < b.N; i++ {
		g.Execute(ctx)
	}
}

var jsonSource string = `{
	"name1": "value1",
	"name2": 1234,
	"name3": true,
	"name4": ["slice elem1", "slice elem2"],
	"my_json_exp_group": [
		[
			["$rand", ">", 5],
			["$my_var", "=", "hello world{{$hour}}"]
		],
		[
			["$rand", "<=", 5],
			[
				["$my_var", "=", "hello json exp-{{$hour}}-"],
				["$my_var", "+=", " hello go lang"]
			]
		],
		[
			["$my_var|len", ">", 10  ],
			["$myobj.version","=","v1.0.rand:{{$rand}}.time:{{$time}}"]
		],
		[
			["$my_var|len", "<=", 10  ],
			["$myobj.version","=","v2.0"]
		]
	]
}`
