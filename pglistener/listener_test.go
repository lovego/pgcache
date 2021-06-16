package pglistener_test

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/lovego/errs"
	loggerPkg "github.com/lovego/logger"
	"github.com/lovego/pgcache/pglistener"
)

var dbUrl = "postgres://postgres:postgres@localhost/postgres?sslmode=disable"
var testDB = connectDB(dbUrl)
var logger = loggerPkg.New(os.Stderr)

type testHandler struct {
}

func (h testHandler) Init(table string) {
	fmt.Printf("Init %s\n", table)
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

func (h testHandler) ConnLoss(table string) {
	fmt.Printf("ConnLoss %s\n", table)
}

func ExampleListener_Listen() {
	testCreateUpdateDelete("students2")
	testCreateUpdateDelete("public.students2")

	// Output:
	// Init public.students2
	// Create public.students2
	//   {"id": 1, "name": "李雷", "time": "2018-09-08"}
	// Update public.students2
	//   old: {"id": 1, "name": "李雷", "time": "2018-09-08"}
	//   new: {"id": 1, "name": "韩梅梅", "time": "2018-09-09"}
	// Delete public.students2
	//   {"id": 1, "name": "韩梅梅", "time": "2018-09-09"}
	// Init public.students2
	// Create public.students2
	//   {"id": 1, "name": "李雷", "time": "2018-09-08"}
	// Update public.students2
	//   old: {"id": 1, "name": "李雷", "time": "2018-09-08"}
	//   new: {"id": 1, "name": "韩梅梅", "time": "2018-09-09"}
	// Delete public.students2
	//   {"id": 1, "name": "韩梅梅", "time": "2018-09-09"}
}

func testCreateUpdateDelete(table string) {
	createStudentsTable()

	listener, err := pglistener.New(dbUrl, nil, logger)
	if err != nil {
		fmt.Println(errs.WithStack(err))
		return
	}
	if err := listener.Listen(
		table,
		"$1.id, $1.name, to_char($1.time, 'YYYY-MM-DD') as time", "",
		testHandler{},
	); err != nil {
		panic(errs.WithStack(err))
	}

	// from now on, testHandler will be notified on INSERT / UPDATE / DELETE.
	if _, err := testDB.Exec(`
    INSERT INTO students2(name, time) VALUES ('李雷', '2018-09-08 15:55:00+08')
  `); err != nil {
		panic(err)
	}
	if _, err = testDB.Exec(`
    UPDATE students2 SET name = '韩梅梅', time = '2018-09-09 15:56:00+08'
  `); err != nil {
		panic(err)
	}
	// this one should not be notified
	if _, err = testDB.Exec(`
    UPDATE students2 SET time = '2018-09-09 15:57:00+08'
  `); err != nil {
		panic(err)
	}
	if _, err = testDB.Exec(`DELETE FROM students2`); err != nil {
		panic(err)
	}

	time.Sleep(10 * time.Millisecond)
	if err := listener.Unlisten(table); err != nil {
		panic(err)
	}
}

func createStudentsTable() {
	if _, err := testDB.Exec(`
	DROP TABLE IF EXISTS students2;
	CREATE TABLE IF NOT EXISTS students2 (
		id   bigserial,
		name varchar(100),
		time timestamptz,
    other text default ''
	)`); err != nil {
		panic(err)
	}
}

func connectDB(dbUrl string) *sql.DB {
	db, err := sql.Open(`postgres`, dbUrl)
	if err != nil {
		panic(err)
	}
	return db
}
