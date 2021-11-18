// Copyright 2021 fangyousong(方友松). All rights reserved.

package mysqlutil

import (
	"fmt"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type ConnPair struct {
	dbS      string
	dbUnsafe *sqlx.DB
	dbSafe   *sqlx.DB
}

type BusyPool struct {
	q []ConnPair
}

type Pool struct {
	lock        sync.Mutex
	connections map[string]*[]ConnPair
}

var dbPool *Pool = nil
var busyPool *BusyPool

func init() {
	dbPool = new(Pool)
	dbPool.connections = make(map[string]*[]ConnPair)
	busyPool = new(BusyPool)
	busyPool.q = make([]ConnPair, 0, 100)
}

//获取数据库连接
func AcquireDb(dsn string) (*sqlx.DB, error) {
	var ret *sqlx.DB
	var err error
	for i := 0; i < 3; i++ {
		ret, err = acquireDb(dsn)
		if ret == nil {
			continue
		}
		if pingDb(ret) {
			return ret, err
		} else {
			ReleaseDb(ret)
		}
	}

	return nil, err
}
func acquireDb(dsn string) (*sqlx.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("dsn is blank")
	}
	dbPool.lock.Lock()
	defer dbPool.lock.Unlock()
	connList, ok := dbPool.connections[dsn]
	if !ok {
		lst := make([]ConnPair, 0, 36)
		connList = &lst
		dbPool.connections[dsn] = connList
	}
	if len(*connList) == 0 {
		sdb, err := sqlx.Connect("mysql", dsn)
		if err != nil {
			return nil, fmt.Errorf("connect mysql fail, %s", err.Error())
		}
		db := sdb.Unsafe()
		busyPool.q = append(busyPool.q, ConnPair{dbSafe: sdb, dbUnsafe: db, dbS: dsn})
		return db, nil
	} else {
		pair := (*connList)[len(*connList)-1]
		*connList = (*connList)[:len(*connList)-1]
		busyPool.q = append(busyPool.q, pair)
		return pair.dbUnsafe, nil
	}
}

func pingDb(conn *sqlx.DB) bool {
	tm := make([]uint8, 0)
	if err := conn.Get(&tm, "select now()"); err == nil {
		return true
	}
	return false
}

//连接放回pool
func ReleaseDb(conn *sqlx.DB) {
	if conn == nil {
		return
	}
	invalidConn := !pingDb(conn)
	dbPool.lock.Lock()
	defer dbPool.lock.Unlock()
	q := make([]ConnPair, 0, 100)
	for _, v := range busyPool.q {
		if v.dbUnsafe != conn {
			q = append(q, v)
		} else {
			if invalidConn {
				v.dbUnsafe.Close()
				v.dbSafe.Close()
			} else {
				connList, ok := dbPool.connections[v.dbS]
				if !ok {
					v.dbUnsafe.Close()
					v.dbSafe.Close()
				} else {
					*connList = append(*connList, v)
				}
			}
		}
	}
	busyPool.q = q
}
