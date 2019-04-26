package pgnotify

import (
	"database/sql"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/lovego/bsql"
	"github.com/lovego/errs"
	loggerPkg "github.com/lovego/logger"
	"github.com/lovego/pgnotify/pghandler"
)

type StudentRow struct {
	Id         int64
	Name, Time string
}

func ExampleNotifier_Notify() {
	db, err := sql.Open(`postgres`, getTestDataSource())
	if err != nil {
		panic(err)
	}

	var mutex sync.RWMutex
	var m map[int64]StudentRow
	var logger = loggerPkg.New(os.Stderr)
	var handler = pghandler.New(pghandler.Table{
		Name:         "students",
		Columns:      "$1.id, $1.name, to_char($1.time, 'YYYY-MM-DD') as time",
		CheckColumns: "$1.id, $1.name",
	}, StudentRow{}, []pghandler.Data{
		{RWMutex: &mutex, MapPtr: &m, MapKeys: []string{"Id"}},
	}, bsql.New(db, time.Second), logger)

	createStudentsTable(db)
	notifier, err := New(getTestDataSource(), logger)
	if err != nil {
		fmt.Println(errs.WithStack(err))
		return
	}
	if err := notifier.Notify(handler); err != nil {
		panic(errs.WithStack(err))
	}

	fmt.Println(m)

	if _, err := db.Exec(`
    INSERT INTO students(name, time) VALUES ('李雷', '2018-09-08 15:55:00+08')
  `); err != nil {
		panic(err)
	}
	time.Sleep(10 * time.Millisecond)
	fmt.Println(m)

	if _, err = db.Exec(`
    UPDATE students SET name = '韩梅梅', time = '2018-09-09 15:56:00+08'
  `); err != nil {
		panic(err)
	}
	time.Sleep(10 * time.Millisecond)
	fmt.Println(m)

	// should not notify
	if _, err = db.Exec(`
    UPDATE students SET time = '2018-09-10 15:57:00+08'
  `); err != nil {
		panic(err)
	}
	time.Sleep(10 * time.Millisecond)
	fmt.Println(m)

	if _, err = db.Exec(`DELETE FROM students`); err != nil {
		panic(err)
	}

	time.Sleep(10 * time.Millisecond)
	fmt.Println(m)

	// Output:
	// map[]
	// map[1:{1 李雷 2018-09-08}]
	// map[1:{1 韩梅梅 2018-09-09}]
	// map[1:{1 韩梅梅 2018-09-09}]
	// map[]
}
