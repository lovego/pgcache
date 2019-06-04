package cache

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/lovego/maps"
)

type Score struct {
	StudentId int
	Subject   string
	Score     int
}

func (s Score) Valid() bool {
	return s.Score >= 0
}

func (s *Score) Valid2() bool {
	return s.Score >= 0
}

func (s *Score) Other() {
}

func ExampleData_precond_1() {
	mutex := sync.RWMutex{}
	var m map[int]Score
	d := Data{
		RWMutex: &mutex,
		MapPtr:  &m, MapKeys: []string{"StudentId"},
		Precond: "Valid",
	}
	d.init(reflect.TypeOf(Score{}))
	fmt.Println(
		d.precond(reflect.ValueOf(&Score{Score: 1}).Elem()),
		d.precond(reflect.ValueOf(&Score{Score: -1}).Elem()),
	)
	// Output: true false
}

func ExampleData_precond_2() {
	mutex := sync.RWMutex{}
	var m map[int]Score
	d := Data{
		RWMutex: &mutex,
		MapPtr:  &m, MapKeys: []string{"StudentId"},
		Precond: "Valid2",
	}
	d.init(reflect.TypeOf(Score{}))
	fmt.Println(
		d.precond(reflect.ValueOf(&Score{Score: 1}).Elem()),
		d.precond(reflect.ValueOf(&Score{Score: -1}).Elem()),
	)
	// Output: true false
}

func ExampleData_save_remove_clear() {
	mutex := sync.RWMutex{}
	var m map[int]int
	d := Data{
		RWMutex: &mutex,
		MapPtr:  &m, MapKeys: []string{"StudentId"}, MapValue: "Score",
		Precond: "Valid",
	}
	d.init(reflect.TypeOf(Score{}))
	d.clear()

	rows := reflect.ValueOf([]Score{
		{StudentId: 1001, Score: 98},
		{StudentId: 1002, Score: 101},
		{StudentId: 1003, Score: 99},
		{StudentId: 1002, Score: 100},
	})
	for i := 0; i < rows.Len(); i++ {
		d.save(rows.Index(i))
	}
	maps.Println(m)

	rows = reflect.ValueOf([]Score{
		{StudentId: 1002, Score: 100},
		{StudentId: 1004},
	})
	for i := 0; i < rows.Len(); i++ {
		d.remove(rows.Index(i))
	}
	maps.Println(m)

	d.clear()
	maps.Println(m)

	// Output:
	// map[1001:98 1002:100 1003:99]
	// map[1001:98 1003:99]
	// map[]
}

func ExampleData_save_remove_clear_sorted_sets() {
	mutex := sync.RWMutex{}
	var m map[int][]int
	d := Data{
		RWMutex: &mutex,
		MapPtr:  &m, MapKeys: []string{"StudentId"}, MapValue: "Score",
		Precond: "Valid",
	}
	d.init(reflect.TypeOf(Score{}))
	rows := reflect.ValueOf([]Score{
		{StudentId: 1001, Score: 98},
		{StudentId: 1001, Score: 99},
		{StudentId: 1002, Score: 90},
		{StudentId: 1003, Score: 99},
		{StudentId: 1002, Score: 91},
		{StudentId: 1003, Score: 100},
		{StudentId: 1001, Score: 99},
	})
	for i := 0; i < rows.Len(); i++ {
		d.save(rows.Index(i))
	}
	maps.Println(m)

	rows = reflect.ValueOf([]Score{
		{StudentId: 1002, Score: 90},
		{StudentId: 1002, Score: 91},
		{StudentId: 1003, Score: 100},
		{StudentId: 1004},
	})
	for i := 0; i < rows.Len(); i++ {
		d.remove(rows.Index(i))
	}
	maps.Println(m)

	d.clear()
	maps.Println(m)

	// Output:
	// map[1001:[98 99] 1002:[90 91] 1003:[99 100]]
	// map[1001:[98 99] 1003:[99]]
	// map[]
}

func ExampleData_save_remove_clear_sorted_sets_2() {
	mutex := sync.RWMutex{}
	var m map[int][]Score
	d := Data{
		RWMutex: &mutex,
		MapPtr:  &m, MapKeys: []string{"StudentId"}, SortedSetUniqueKey: []string{"Subject"},
	}
	d.init(reflect.TypeOf(Score{}))
	rows := reflect.ValueOf([]Score{
		{StudentId: 1001, Subject: "语文", Score: 98},
		{StudentId: 1001, Subject: "语文", Score: 99},
		{StudentId: 1002, Subject: "数学", Score: 90},
		{StudentId: 1003, Subject: "语文", Score: 99},
		{StudentId: 1002, Subject: "数学", Score: 91},
		{StudentId: 1003, Subject: "数学", Score: 100},
	})
	for i := 0; i < rows.Len(); i++ {
		d.save(rows.Index(i))
	}
	maps.Println(m)

	rows = reflect.ValueOf([]Score{
		{StudentId: 1001, Subject: "语文", Score: 90},
		{StudentId: 1002, Subject: "数学", Score: 90},
		{StudentId: 1003, Subject: "数学"},
		{StudentId: 1004, Subject: "语文"},
	})
	for i := 0; i < rows.Len(); i++ {
		d.remove(rows.Index(i))
	}
	maps.Println(m)

	d.clear()
	maps.Println(m)

	// Output:
	// map[1001:[{1001 语文 99}] 1002:[{1002 数学 91}] 1003:[{1003 数学 100} {1003 语文 99}]]
	// map[1003:[{1003 语文 99}]]
	// map[]
}

func ExampleData_save_remove_clear_sorted_sets_3() {
	mutex := sync.RWMutex{}
	var m map[int]map[string][]int
	d := Data{
		RWMutex: &mutex,
		MapPtr:  &m, MapKeys: []string{"StudentId", "Subject"}, MapValue: "Score",
	}
	d.init(reflect.TypeOf(Score{}))
	rows := reflect.ValueOf([]Score{
		{StudentId: 1001, Subject: "语文", Score: 98},
		{StudentId: 1001, Subject: "语文", Score: 99},
		{StudentId: 1002, Subject: "数学", Score: 91},
		{StudentId: 1003, Subject: "语文", Score: 99},
		{StudentId: 1002, Subject: "数学", Score: 90},
		{StudentId: 1003, Subject: "数学", Score: 100},
	})
	for i := 0; i < rows.Len(); i++ {
		d.save(rows.Index(i))
	}
	maps.Println(m)

	rows = reflect.ValueOf([]Score{
		{StudentId: 1002, Subject: "数学", Score: 90},
		{StudentId: 1002, Subject: "数学", Score: 91},
		{StudentId: 1004, Subject: "语文"},
	})
	for i := 0; i < rows.Len(); i++ {
		d.remove(rows.Index(i))
	}
	maps.Println(m)

	m = make(map[int]map[string][]int)
	fmt.Println(d.mapV)

	// Output:
	// map[1001:map[语文:[98 99]] 1002:map[数学:[90 91]] 1003:map[数学:[100] 语文:[99]]]
	// map[1001:map[语文:[98 99]] 1002:map[] 1003:map[数学:[100] 语文:[99]]]
	// map[]
}

func ExampleAddKeyValueNames() {
	var m map[string]map[int64]*uint16
	src := reflect.TypeOf(m).String()
	fmt.Println(addKeyValueNames(src, nil, ``))
	fmt.Println(addKeyValueNames(src, []string{`Type`}, ``))
	fmt.Println(addKeyValueNames(src, []string{`Type`, `Id`}, ``))
	fmt.Println(addKeyValueNames(src, []string{`Type`, `Id`, `XXX`}, `Flags`))
	// Output:
	// map[string]map[int64]*uint16
	// map[Type:string]map[int64]*uint16
	// map[Type:string]map[Id:int64]*uint16
	// map[Type:string]map[Id:int64]Flags:*uint16
}
