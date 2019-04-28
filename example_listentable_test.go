package pglistener_test

import (
	"fmt"
	"sync"
	"time"

	"github.com/lovego/bsql"
	"github.com/lovego/errs"
	"github.com/lovego/pglistener"
	"github.com/lovego/pglistener/cache"
)

type StudentRow struct {
	Id         int64
	Name, Time string
}

func getStudentsCacheHandler(mapPtr interface{}) *cache.Handler {
	var mutex sync.RWMutex
	return cache.New(cache.Table{
		Name:         "students",
		Columns:      "$1.id, $1.name, to_char($1.time, 'YYYY-MM-DD') as time",
		CheckColumns: "$1.id, $1.name",
	}, StudentRow{}, []cache.Data{
		{RWMutex: &mutex, MapPtr: mapPtr, MapKeys: []string{"Id"}},
	}, bsql.New(testDB, time.Second), logger)
}

func ExampleListener_ListenTable() {
	createStudentsTable()

	var students map[int64]StudentRow

	listener, err := pglistener.New(dbUrl, logger)
	if err != nil {
		fmt.Println(errs.WithStack(err))
		return
	}
	if err := listener.ListenTable(getStudentsCacheHandler(&students)); err != nil {
		fmt.Println(errs.WithStack(err))
		return
	}

	fmt.Println(students)

	if _, err := testDB.Exec(`
    INSERT INTO students(name, time) VALUES ('李雷', '2018-09-08 15:55:00+08')
  `); err != nil {
		panic(err)
	}
	time.Sleep(10 * time.Millisecond)
	fmt.Println(students)

	if _, err = testDB.Exec(`
    UPDATE students SET name = '韩梅梅', time = '2018-09-09 15:56:00+08'
  `); err != nil {
		panic(err)
	}
	time.Sleep(10 * time.Millisecond)
	fmt.Println(students)

	// should not be notified
	if _, err = testDB.Exec(`
    UPDATE students SET time = '2018-09-10 15:57:00+08'
  `); err != nil {
		panic(err)
	}
	time.Sleep(10 * time.Millisecond)
	fmt.Println(students)

	if _, err = testDB.Exec(`DELETE FROM students`); err != nil {
		panic(err)
	}

	time.Sleep(10 * time.Millisecond)
	fmt.Println(students)

	// Output:
	// map[]
	// map[1:{1 李雷 2018-09-08}]
	// map[1:{1 韩梅梅 2018-09-09}]
	// map[1:{1 韩梅梅 2018-09-09}]
	// map[]
}
