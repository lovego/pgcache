# pglistener
- Listen for INSERT/UPDATE/DELETE events of postgresql's table, and pass the events to a defined handler.

[![Build Status](https://travis-ci.org/lovego/pglistener.svg?branch=master)](https://travis-ci.org/lovego/pglistener)
[![Coverage Status](https://img.shields.io/coveralls/github/lovego/pglistener/master.svg)](https://coveralls.io/github/lovego/pglistener?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/lovego/pglistener)](https://goreportcard.com/report/github.com/lovego/pglistener)
[![GoDoc](https://godoc.org/github.com/lovego/pglistener?status.svg)](https://godoc.org/github.com/lovego/pglistener)

## Install
`$ go get github.com/lovego/pglistener`

## Example (cache a table)
```go
package pglistener_test

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
    "github.com/lovego/pglistener"
    "github.com/lovego/pglistener/cache"
)

var dbUrl = "postgres://postgres:@localhost/test?sslmode=disable"
var testDB = connectDB(dbUrl)
var logger = loggerPkg.New(os.Stderr)

type Student struct {
    Id    int64
    Name  string
    Class string
}

func ExampleListener() {
    initStudentsTable()

    var studentsMap = make(map[int64]Student)
    var classesMap = make(map[string][]Student)

    listener, err := pglistener.New(dbUrl, logger)
    if err != nil {
        panic(err)
    }
    if err := listener.ListenTable(getTableHandler(&studentsMap, &classesMap)); err != nil {
        panic(err)
    }

    // from now on, studentsMap and classesMap is always synchronized with students table.
    fmt.Println(`init:`)
    maps.Println(studentsMap)
    maps.Println(classesMap)

    // even you insert some rows.
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

    // even you update some rows.
    if _, err := testDB.Exec(`UPDATE students SET class = '初三2班'`); err != nil {
        panic(err)
    }
    time.Sleep(10 * time.Millisecond)
    fmt.Println(`after UPDATE:`)
    maps.Println(studentsMap)
    maps.Println(classesMap)

    // even you delete some rows.
    if _, err := testDB.Exec(`DELETE FROM students WHERE id in (3, 4)`); err != nil {
        panic(err)
    }
    time.Sleep(10 * time.Millisecond)
    fmt.Println(`after DELETE:`)
    maps.Println(studentsMap)
    maps.Println(classesMap)

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

func getTableHandler(studentsMap, classesMap interface{}) pglistener.TableHandler {
    var mutex sync.RWMutex

    return cache.New(cache.Table{Name: "students"}, Student{}, []cache.Data{
        {
            RWMutex: &mutex, MapPtr: studentsMap, MapKeys: []string{"Id"},
        }, {
            RWMutex: &mutex, MapPtr: classesMap, MapKeys: []string{"Class"},
            SortedSetUniqueKey: []string{"Id"},
        },
    }, bsql.New(testDB, time.Second), logger)
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
[https://godoc.org/github.com/lovego/pglistener](https://godoc.org/github.com/lovego/pglistener)

## Test
`PG_DATA_SOURCE="postgres://user:password@host:port/db?sslmode=disable" go test`

