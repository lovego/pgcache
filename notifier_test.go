package pgnotify

import (
	"database/sql"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/lovego/errs"
	"github.com/lovego/logger"
)

type testHandler struct {
}

func (h testHandler) ConnLoss(table string) {
	fmt.Printf("ConnLoss %s\n", table)
}

func (h testHandler) Create(table string, newBuf []byte) {
	fmt.Printf("Create %s\n  %s\n", table, newBuf)
}

func (h testHandler) Update(table string, oldBuf, newBuf []byte) {
	fmt.Printf("Update %s\n  old: %s\n  new: %s\n", table, oldBuf, newBuf)
}

func (h testHandler) Delete(table string, oldBuf []byte) {
	fmt.Printf("Delete %s\n  %s\n", table, oldBuf)
}

func ExampleNotifier() {
	db, err := sql.Open(`postgres`, getTestDataSource())
	if err != nil {
		panic(err)
	}
	createStudentsTable(db)

	notifier, err := New(getTestDataSource(), logger.New(os.Stderr))
	if err != nil {
		fmt.Println(errs.WithStack(err))
		return
	}
	if err := notifier.Notify(
		"students",
		"$1.id, $1.name, to_char($1.time, 'YYYY-MM-DD') as time",
		"$1.id, $1.name",
		testHandler{},
	); err != nil {
		panic(errs.WithStack(err))
	}

	if _, err := db.Exec(`
    INSERT INTO students(name, time) VALUES ('李雷', '2018-09-08 15:55:00+08')
  `); err != nil {
		panic(err)
	}
	if _, err = db.Exec(`
    UPDATE students SET name = '韩梅梅', time = '2018-09-09 15:56:00+08'
  `); err != nil {
		panic(err)
	}
	// should not notify
	if _, err = db.Exec(`
    UPDATE students SET time = '2018-09-10 15:57:00+08'
  `); err != nil {
		panic(err)
	}
	if _, err = db.Exec(`DELETE FROM students`); err != nil {
		panic(err)
	}

	time.Sleep(10 * time.Millisecond)

	// Output:
	// ConnLoss students
	// Create students
	//   {"id": 1, "name": "李雷", "time": "2018-09-08"}
	// Update students
	//   old: {"id": 1, "name": "李雷", "time": "2018-09-08"}
	//   new: {"id": 1, "name": "韩梅梅", "time": "2018-09-09"}
	// Delete students
	//   {"id": 1, "name": "韩梅梅", "time": "2018-09-10"}

}

func createStudentsTable(db *sql.DB) {
	if _, err := db.Exec(`
	DROP TABLE IF EXISTS students;
	CREATE TABLE IF NOT EXISTS students (
		id   bigserial,
		name varchar(100),
		time timestamptz,
    other text default 'not to notify'
	)`); err != nil {
		panic(err)
	}
}

func getTestDataSource() string {
	if env := os.Getenv("PG_DATA_SOURCE"); env != "" {
		return env
	} else if runtime.GOOS == "darwin" {
		return "postgres://postgres:@localhost/test?sslmode=disable"
	} else {
		return "postgres://travis:123456@localhost:5433/travis?sslmode=disable"
	}
}
