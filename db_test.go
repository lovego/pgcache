package pgcache_test

import (
	"database/sql"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/lovego/bsql"
	loggerPkg "github.com/lovego/logger"
	"github.com/lovego/maps"
	"github.com/lovego/pgcache"
)

var dbUrl = "postgres://postgres:@localhost/travis?sslmode=disable"
var testDB = connectDB(dbUrl)
var logger = loggerPkg.New(os.Stderr)

type Student struct {
	Id        int64
	Name      string
	Class     string
	UpdatedAt time.Time `json:"updated_at"`
}

func (s Student) String() string {
	return fmt.Sprintf(`{%d %s %s %s}`, s.Id, s.Name, s.Class,
		s.UpdatedAt.Format(`2006-01-02 15:04:05 Z0700`))
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
	// map[1:{1 李雷 初三1班 2003-10-01 09:10:10 +0800} 2:{2 韩梅梅 初三1班 2003-10-01 09:10:20 +0800}]
	// map[初三1班:[{1 李雷 初三1班 2003-10-01 09:10:10 +0800} {2 韩梅梅 初三1班 2003-10-01 09:10:20 +0800}]]
	// after INSERT:
	// map[1:{1 李雷 初三1班 2003-10-01 09:10:10 +0800} 2:{2 韩梅梅 初三1班 2003-10-01 09:10:20 +0800} 3:{3 Lily 初三2班 2003-10-01 09:10:30 +0800} 4:{4 Lucy 初三2班 2003-10-01 09:10:30 +0800}]
	// map[初三1班:[{1 李雷 初三1班 2003-10-01 09:10:10 +0800} {2 韩梅梅 初三1班 2003-10-01 09:10:20 +0800}] 初三2班:[{3 Lily 初三2班 2003-10-01 09:10:30 +0800} {4 Lucy 初三2班 2003-10-01 09:10:30 +0800}]]
	// after UPDATE:
	// map[1:{1 李雷 初三2班 2003-10-01 09:10:40 +0800} 2:{2 韩梅梅 初三2班 2003-10-01 09:10:40 +0800} 3:{3 Lily 初三2班 2003-10-01 09:10:40 +0800} 4:{4 Lucy 初三2班 2003-10-01 09:10:40 +0800}]
	// map[初三2班:[{1 李雷 初三2班 2003-10-01 09:10:40 +0800} {2 韩梅梅 初三2班 2003-10-01 09:10:40 +0800} {3 Lily 初三2班 2003-10-01 09:10:40 +0800} {4 Lucy 初三2班 2003-10-01 09:10:40 +0800}]]
	// after DELETE:
	// map[1:{1 李雷 初三2班 2003-10-01 09:10:40 +0800} 2:{2 韩梅梅 初三2班 2003-10-01 09:10:40 +0800}]
	// map[初三2班:[{1 李雷 初三2班 2003-10-01 09:10:40 +0800} {2 韩梅梅 初三2班 2003-10-01 09:10:40 +0800}]]
}

func initStudentsTable() {
	if _, err := testDB.Exec(`
SET TIME ZONE 'PRC';
DROP TABLE IF EXISTS students;
CREATE TABLE IF NOT EXISTS students (
  id    bigserial,
  name  text,
  class text,
  updated_at timestamptz
);
INSERT INTO students (id, name, class, updated_at)
VALUES
(1, '李雷',   '初三1班', '2003-10-1T09:10:10+08:00'),
(2, '韩梅梅', '初三1班', '2003-10-1T09:10:20+08:00');
`); err != nil {
		panic(err)
	}
}

func testInsert(studentsMap map[int64]Student, classesMap map[string][]Student) {
	if _, err := testDB.Exec(`
INSERT INTO students (id, name, class, updated_at)
VALUES
(3, 'Lily',   '初三2班', '2003-10-1T09:10:30+08:00'),
(4, 'Lucy',   '初三2班', '2003-10-1T09:10:30+08:00');
`); err != nil {
		panic(err)
	}
	time.Sleep(10 * time.Millisecond)
	fmt.Println(`after INSERT:`)
	maps.Println(studentsMap)
	maps.Println(classesMap)
}

func testUpdate(studentsMap map[int64]Student, classesMap map[string][]Student) {
	if _, err := testDB.Exec(
		`UPDATE students SET "class" = '初三2班', updated_at = '2003-10-1 09:10:40+08:00'`,
	); err != nil {
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

func Example2() {
	initStudentsTable()

	var studentsMap = make(map[int64]Student)
	var classesMap = make(map[string][]Student)
	var studentsSlice = make([]Student, 0, 4)
	var mutex sync.RWMutex

	dbCache, err := pgcache.New(dbUrl, bsql.New(testDB, time.Second), logger)
	if err != nil {
		panic(err)
	}
	tableCache, err := dbCache.Add(&pgcache.Table{
		Name:       "students",
		RowStruct:  Student{},
		BigColumns: "name",
		Datas: []*pgcache.Data{
			{
				RWMutex: &mutex, DataPtr: &studentsMap, MapKeys: []string{"Id"},
			}, {
				RWMutex: &mutex, DataPtr: &classesMap, MapKeys: []string{"Class"},
				SortedSetUniqueKey: []string{"Id"},
			}, {
				RWMutex: &mutex, DataPtr: &studentsSlice, SortedSetUniqueKey: []string{"Id"},
			},
		},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(tableCache.Columns, tableCache.BigColumns)
	fmt.Println(tableCache.LoadSql)

	// from now on, studentsMap and classesMap is always synchronized with students table.
	fmt.Println(`init:`)
	maps.Println(studentsMap)
	maps.Println(classesMap)
	maps.Println(studentsSlice)

	// even you insert some rows.
	testInsert(studentsMap, classesMap)
	maps.Println(studentsSlice)
	// even you update some rows.
	testUpdate(studentsMap, classesMap)
	maps.Println(studentsSlice)
	// even you delete some rows.
	testDelete(studentsMap, classesMap)
	maps.Println(studentsSlice)

	dbCache.RemoveAll()

	// Output:
	// id,class,updated_at name
	// SELECT id,class,updated_at ,name FROM students
	// init:
	// map[1:{1 李雷 初三1班 2003-10-01 09:10:10 +0800} 2:{2 韩梅梅 初三1班 2003-10-01 09:10:20 +0800}]
	// map[初三1班:[{1 李雷 初三1班 2003-10-01 09:10:10 +0800} {2 韩梅梅 初三1班 2003-10-01 09:10:20 +0800}]]
	// [{1 李雷 初三1班 2003-10-01 09:10:10 +0800} {2 韩梅梅 初三1班 2003-10-01 09:10:20 +0800}]
	// after INSERT:
	// map[1:{1 李雷 初三1班 2003-10-01 09:10:10 +0800} 2:{2 韩梅梅 初三1班 2003-10-01 09:10:20 +0800} 3:{3 Lily 初三2班 2003-10-01 09:10:30 +0800} 4:{4 Lucy 初三2班 2003-10-01 09:10:30 +0800}]
	// map[初三1班:[{1 李雷 初三1班 2003-10-01 09:10:10 +0800} {2 韩梅梅 初三1班 2003-10-01 09:10:20 +0800}] 初三2班:[{3 Lily 初三2班 2003-10-01 09:10:30 +0800} {4 Lucy 初三2班 2003-10-01 09:10:30 +0800}]]
	// [{1 李雷 初三1班 2003-10-01 09:10:10 +0800} {2 韩梅梅 初三1班 2003-10-01 09:10:20 +0800} {3 Lily 初三2班 2003-10-01 09:10:30 +0800} {4 Lucy 初三2班 2003-10-01 09:10:30 +0800}]
	// after UPDATE:
	// map[1:{1 李雷 初三2班 2003-10-01 09:10:40 +0800} 2:{2 韩梅梅 初三2班 2003-10-01 09:10:40 +0800} 3:{3 Lily 初三2班 2003-10-01 09:10:40 +0800} 4:{4 Lucy 初三2班 2003-10-01 09:10:40 +0800}]
	// map[初三2班:[{1 李雷 初三2班 2003-10-01 09:10:40 +0800} {2 韩梅梅 初三2班 2003-10-01 09:10:40 +0800} {3 Lily 初三2班 2003-10-01 09:10:40 +0800} {4 Lucy 初三2班 2003-10-01 09:10:40 +0800}]]
	// [{1 李雷 初三2班 2003-10-01 09:10:40 +0800} {2 韩梅梅 初三2班 2003-10-01 09:10:40 +0800} {3 Lily 初三2班 2003-10-01 09:10:40 +0800} {4 Lucy 初三2班 2003-10-01 09:10:40 +0800}]
	// after DELETE:
	// map[1:{1 李雷 初三2班 2003-10-01 09:10:40 +0800} 2:{2 韩梅梅 初三2班 2003-10-01 09:10:40 +0800}]
	// map[初三2班:[{1 李雷 初三2班 2003-10-01 09:10:40 +0800} {2 韩梅梅 初三2班 2003-10-01 09:10:40 +0800}]]
	// [{1 李雷 初三2班 2003-10-01 09:10:40 +0800} {2 韩梅梅 初三2班 2003-10-01 09:10:40 +0800}]
}

func connectDB(dbUrl string) *sql.DB {
	db, err := sql.Open(`postgres`, dbUrl)
	if err != nil {
		panic(err)
	}
	return db
}
