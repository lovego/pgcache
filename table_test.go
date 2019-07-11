package pgcache

import (
	"database/sql"
	"fmt"
	"os"
	"reflect"
	"sync"

	"github.com/lovego/logger"
	"github.com/lovego/maps"
)

var testLogger = logger.New(os.Stdout)

type testQuerier struct{}

func (q testQuerier) Query(data interface{}, sql string, args ...interface{}) error {
	rows := data.(*[]Score)
	*rows = []Score{
		{StudentId: 1000, Subject: "语文", Score: 90},
	}
	return nil
}

func (q testQuerier) GetDB() *sql.DB {
	return nil
}

func ExampleTable() {
	var m1 map[int]map[string]int
	var m2 map[string]map[int]int

	var mutex sync.RWMutex
	t := &Table{
		Name:      "scores",
		RowStruct: Score{},
		Datas: []*Data{
			{RWMutex: &mutex, DataPtr: &m1, MapKeys: []string{"StudentId", "Subject"}, Value: "Score"},
			{RWMutex: &mutex, DataPtr: &m2, MapKeys: []string{"Subject", "StudentId"}, Value: "Score"},
		},
	}
	t.init("db", testQuerier{}, testLogger)

	t.Init("")
	maps.Println(m1, m2)

	t.Create("", []byte(`{"StudentId": 1001, "Subject": "语文", "Score": 95}`))
	maps.Println(m1, m2)

	t.Update("",
		[]byte(`{"StudentId": 1001, "Subject": "语文", "Score": 95}`),
		[]byte(`{"StudentId": 1001, "Subject": "数学", "Score": 96}`),
	)
	maps.Println(m1, m2)

	t.Delete("",
		[]byte(`{"StudentId": 1001, "Subject": "数学", "Score": 96}`),
	)
	maps.Println(m1, m2)

	t.Clear()
	maps.Println(m1, m2)

	t.Create("", []byte(`{"StudentId": 1001, "Subject": "语文", "Score": 95}`))
	maps.Println(m1, m2)

	// Output:
	// map[1000:map[语文:90]] map[语文:map[1000:90]]
	// map[1000:map[语文:90] 1001:map[语文:95]] map[语文:map[1000:90 1001:95]]
	// map[1000:map[语文:90] 1001:map[数学:96]] map[数学:map[1001:96] 语文:map[1000:90]]
	// map[1000:map[语文:90] 1001:map[]] map[数学:map[] 语文:map[1000:90]]
	// map[] map[]
	// map[1001:map[语文:95]] map[语文:map[1001:95]]
}

func ExamplePointerValue_1() {
	var m map[string]int
	v := reflect.ValueOf(&m).Elem()
	m = map[string]int{"a": 1}
	fmt.Println(m, v, &m == v.Addr().Interface())
	v.Set(reflect.MakeMap(v.Type()))
	fmt.Println(m, v, &m == v.Addr().Interface())
	v.SetMapIndex(reflect.ValueOf("b"), reflect.ValueOf(2))
	fmt.Println(m, v, &m == v.Addr().Interface())

	// Output:
	// map[a:1] map[a:1] true
	// map[] map[] true
	// map[b:2] map[b:2] true
}

func ExamplePointerValue_2() {
	var m map[string]int
	v := reflect.ValueOf(&m).Elem()
	v.Set(reflect.MakeMap(v.Type()))
	fmt.Println(m, v, &m == v.Addr().Interface())
	v.SetMapIndex(reflect.ValueOf("b"), reflect.ValueOf(2))
	fmt.Println(m, v, &m == v.Addr().Interface())
	m = map[string]int{"a": 1}
	fmt.Println(m, v, &m == v.Addr().Interface())

	// Output:
	// map[] map[] true
	// map[b:2] map[b:2] true
	// map[a:1] map[a:1] true
}
