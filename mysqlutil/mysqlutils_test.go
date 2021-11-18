// Copyright 2021 fangyousong(方友松). All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package mysqlutil

import (
	"encoding/json"
	"fmt"
	"testing"
)

type TestDbObj struct {
	Id    int     `json:"id" db:"id"`
	Name  string  `json:"name" db:"name"`
	Price float64 `json:"price" db:"price"`
	Age   int     `json:"age" db:"age"`
}

func TestUpsertSql(t *testing.T) {
	testObj := TestDbObj{Id: 1234, Name: "hello mysql", Price: 1234.56}
	sql, err := UpsertSql(&testObj, "test_table", []string{"id"})
	if err != nil {
		t.Fatal(err.Error())
	}
	fmt.Printf("upsertSql: %s\n", sql)
}

func TestInsertSql(t *testing.T) {
	testObj := TestDbObj{Id: 1234, Name: "hello mysql", Price: 1234.56}
	sql, err := InsertSql(&testObj, "test_table")
	if err != nil {
		t.Fatal(err.Error())
	}
	fmt.Printf("insertSql: %s\n", sql)
}

func TestDeleteSql(t *testing.T) {
	testObj := TestDbObj{Id: 1234, Name: "hello mysql", Price: 1234.56, Age: 789}
	sql, err := DeleteSql(&testObj, "test_table")
	if err != nil {
		t.Fatal(err.Error())
	}
	fmt.Printf("deleteSql: %s\n", sql)
}

func TestSelectSql(t *testing.T) {
	testObj := TestDbObj{Id: 1234, Name: "hello mysql", Price: 1234.56, Age: 789}
	sql, err := SelectSql(&testObj, "test_table", "")
	if err != nil {
		t.Fatal(err.Error())
	}
	fmt.Printf("selectSql: %s\n", sql)
}

func TestUpsertSqlM(t *testing.T) {
	testObj := TestDbObj{Id: 1234, Name: "hello mysql", Price: 1234.56, Age: 789}
	mp := make(map[string]interface{})
	bts, _ := json.Marshal(&testObj)
	json.Unmarshal(bts, &mp)
	sql, err := UpsertSqlM(mp, "test_table", []string{"id"})
	if err != nil {
		t.Fatal(err.Error())
	}
	fmt.Printf("upsertSqlM: %s\n", sql)
}

func TestInsertSqlM(t *testing.T) {
	testObj := TestDbObj{Id: 1234, Name: "hello mysql", Price: 1234.56, Age: 789}
	mp := make(map[string]interface{})
	bts, _ := json.Marshal(&testObj)
	json.Unmarshal(bts, &mp)
	sql, err := InsertSqlM(mp, "test_table")
	if err != nil {
		t.Fatal(err.Error())
	}
	fmt.Printf("insertSqlM: %s\n", sql)
}

func TestDeleteSqlM(t *testing.T) {
	testObj := TestDbObj{Id: 1234, Name: "hello mysql", Price: 1234.56, Age: 789}
	mp := make(map[string]interface{})
	bts, _ := json.Marshal(&testObj)
	json.Unmarshal(bts, &mp)
	sql, err := DeleteSqlM(mp, "test_table")
	if err != nil {
		t.Fatal(err.Error())
	}
	fmt.Printf("deleteSqlM: %s\n", sql)
}

func TestSelectSqlM(t *testing.T) {
	testObj := TestDbObj{Id: 1234, Name: "hello mysql", Price: 1234.56, Age: 789}
	mp := make(map[string]interface{})
	bts, _ := json.Marshal(&testObj)
	json.Unmarshal(bts, &mp)
	sql, err := SelectSqlM(mp, "test_table", "*")
	if err != nil {
		t.Fatal(err.Error())
	}
	fmt.Printf("selectSqlM: %s\n", sql)
}
