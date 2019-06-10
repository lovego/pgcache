# pgcache
- cache Postgresql table's data in memory and keep the cache up to date.

[![Build Status](https://travis-ci.org/lovego/pgcache.svg?branch=master)](https://travis-ci.org/lovego/pgcache)
[![Coverage Status](https://img.shields.io/coveralls/github/lovego/pgcache/master.svg)](https://coveralls.io/github/lovego/pgcache?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/lovego/pgcache)](https://goreportcard.com/report/github.com/lovego/pgcache)
[![GoDoc](https://godoc.org/github.com/lovego/pgcache?status.svg)](https://godoc.org/github.com/lovego/pgcache)

## Install
`$ go get github.com/lovego/pgcache`

## Example
```go
package pgcache_test

import (
	"database/sql"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/lovego/bsql"
	loggerPkg "github.com/lovego/logger"
	"github.com/lovego/maps"
	"github.com/lovego/pgcache"
)

var dbUrl = "postgres://postgres:@localhost/test?sslmode=disable"
var testDB = connectDB(dbUrl)
var logger = loggerPkg.New(os.Stderr)

type Student struct {
	Id    int64
	Name  string
	Class string
}

func Example() {
	initStudentsTable()

	var studentsMap = make(map[int64]Student)
	var classesMap = make(map[string][]Student)
	var mutex sync.RWMutex

	dbCache, err := pgcache.New(dbUrl, bsql.New(testDB, time.Second), logger)
	if err != nil {
		panic(err)
	}
	_, err = dbCache.Add(&pgcache.Table{
		Name:      "students",
		RowStruct: Student{},
		Datas: []*pgcache.Data{
			{
				RWMutex: &mutex, DataPtr: &studentsMap, MapKeys: []string{"Id"},
			}, {
				RWMutex: &mutex, DataPtr: &classesMap, MapKeys: []string{"Class"},
				SortedSetUniqueKey: []string{"Id"},
			},
		},
	})
	if err != nil {
		panic(err)
	}

	// from now on, studentsMap and classesMap is always synchronized with students table.
	fmt.Println(`init:`)
	maps.Println(studentsMap)
	maps.Println(classesMap)

	// even you insert some rows.
	testInsert(studentsMap, classesMap)
	// even you update some rows.
	testUpdate(studentsMap, classesMap)
	// even you delete some rows.
	testDelete(studentsMap, classesMap)

	dbCache.RemoveAll()

	// Output:
	// init:
	// map[1:{1 李雷 初三1班} 2:{2 韩梅梅 初三1班}]
	// map[初三1班:[{1 李雷 初三1班} {2 韩梅梅 初三1班}]]
	// after INSERT:
	// map[1:{1 李雷 初三1班} 2:{2 韩梅梅 初三1班} 3:{3 Lily 初三2班} 4:{4 Lucy 初三2班}]
	// map[初三1班:[{1 李雷 初三1班} {2 韩梅梅 初三1班}] 初三2班:[{3 Lily 初三2班} {4 Lucy 初三2班}]]
	// after UPDATE:
	// map[1:{1 李雷 初三2班} 2:{2 韩梅梅 初三2班} 3:{3 Lily 初三2班} 4:{4 Lucy 初三2班}]
	// map[初三2班:[{1 李雷 初三2班} {2 韩梅梅 初三2班} {3 Lily 初三2班} {4 Lucy 初三2班}]]
	// after DELETE:
	// map[1:{1 李雷 初三2班} 2:{2 韩梅梅 初三2班}]
	// map[初三2班:[{1 李雷 初三2班} {2 韩梅梅 初三2班}]]
}

func initStudentsTable() {
	if _, err := testDB.Exec(`
DROP TABLE IF EXISTS students;
CREATE TABLE IF NOT EXISTS students (
  id    bigserial,
  name  text,
  class text
);
INSERT INTO students (id, name, class)
VALUES
(1, '李雷',   '初三1班'),
(2, '韩梅梅', '初三1班');
`); err != nil {
		panic(err)
	}
}

func testInsert(studentsMap map[int64]Student, classesMap map[string][]Student) {
	if _, err := testDB.Exec(`
INSERT INTO students (id, name, class)
VALUES
(3, 'Lily',   '初三2班'),
(4, 'Lucy',   '初三2班');
`); err != nil {
		panic(err)
	}
	time.Sleep(10 * time.Millisecond)
	fmt.Println(`after INSERT:`)
	maps.Println(studentsMap)
	maps.Println(classesMap)
}

func testUpdate(studentsMap map[int64]Student, classesMap map[string][]Student) {
	if _, err := testDB.Exec(`UPDATE students SET "class" = '初三2班'`); err != nil {
		panic(err)
	}
	time.Sleep(10 * time.Millisecond)
	fmt.Println(`after UPDATE:`)
	maps.Println(studentsMap)
	maps.Println(classesMap)
}

func testDelete(studentsMap map[int64]Student, classesMap map[string][]Student) {
	if _, err := testDB.Exec(`DELETE FROM students WHERE id in (3, 4)`); err != nil {
		panic(err)
	}
	time.Sleep(10 * time.Millisecond)
	fmt.Println(`after DELETE:`)
	maps.Println(studentsMap)
	maps.Println(classesMap)
}

func connectDB(dbUrl string) *sql.DB {
	db, err := sql.Open(`postgres`, dbUrl)
	if err != nil {
		panic(err)
	}
	return db
}
```

## Docs
[https://godoc.org/github.com/lovego/pgcache](https://godoc.org/github.com/lovego/pgcache)

## Test
`PG_DATA_SOURCE="postgres://user:password@host:port/db?sslmode=disable" go test`

