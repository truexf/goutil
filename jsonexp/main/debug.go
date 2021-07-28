package main

import (
	"fmt"
	"time"

	"github.com/truexf/goutil/jsonexp"
)

func main() {
	dict := jsonexp.NewDictionary()
	dict.RegisterVar("$my_var", nil)
	cfg, err := jsonexp.NewConfiguration([]byte(jsonSource), dict)
	if err != nil {
		fmt.Println(err.Error())
	}
	if v, _ := cfg.GetNameValue("name1", nil); v.(string) != "value1" {
		fmt.Println("name value fail")
	}
	if v, _ := cfg.GetNameValue("name2", nil); int(v.(float64)) != 1234 {
		fmt.Println("name value fail")
	}
	if v, _ := cfg.GetNameValue("name3", nil); v.(bool) {
		fmt.Println("name value fail")
	}
	if v, _ := cfg.GetNameValue("name4", nil); len(v.([]interface{})) != 2 {
		fmt.Println("name value fail")
	}
	// fmt.Printf("%v\n", cfg.)
	g, ok := cfg.GetJsonExpGroup("my_json_exp_group")
	if !ok {
		fmt.Println("no jsonexp group")
	}
	ctx := &jsonexp.DefaultContext{}
	ctx.SetCtxData("$rand", time.Now().Second()%10)
	g.Execute(ctx)
	myVar, _ := ctx.GetCtxData("$my_var")
	fmt.Printf("%v", myVar)
}

var jsonSource string = `{
	"name1": "value1",
	"name2": 1234,
	"name3": true,
	"name4": ["slice elem1", "slice elem2"],
	"my_json_exp_group": [
		[
			["$rand", ">", 5],
			["$my_var", "=", "hello world"]
		],
		[
			["$rand", "<=", 5],
			[
				["$my_var", "=", "hello json exp"],
				["$my_var", "+=", " hello go lang"]
			]
		]
	]
}`
