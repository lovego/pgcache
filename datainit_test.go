package pgcache

import (
	"fmt"
	"reflect"
	"sync"
)

func ExampleData_init_nilMutex() {
	d := Data{}
	fmt.Println(d.init(nil))
	// Output:
	// Data.RWMutex is nil.
}

func ExampleData_init_invalidDataPtr_1() {
	mutex := sync.RWMutex{}

	d := Data{RWMutex: &mutex, DataPtr: map[int]int{}}
	fmt.Println(d.init(nil))
	// Output:
	// Data.DataPtr should be a non nil pointer to a map or slice.
}

func ExampleData_init_invalidDataPtr_2() {
	mutex := sync.RWMutex{}
	var p *map[int]int
	d := Data{RWMutex: &mutex, DataPtr: p}
	fmt.Println(d.init(nil))
	// Output:
	// Data.DataPtr should be a non nil pointer to a map or slice.
}

func ExampleData_init_invalidMapKeys_1() {
	mutex := sync.RWMutex{}
	var m map[int]int
	d := Data{RWMutex: &mutex, DataPtr: &m}
	fmt.Println(d.init(nil))
	// Output:
	// Data.DataPtr is a 1 layers map, but Data.MapKeys has 0 field.
}

func ExampleData_init_invalidMapKeys_2() {
	mutex := sync.RWMutex{}
	var m map[int]map[string]int
	d := Data{RWMutex: &mutex, DataPtr: &m, MapKeys: []string{"StudentId", "Subject", "Other"}}
	fmt.Println(d.init(reflect.TypeOf(Score{})))
	// Output:
	// Data.DataPtr is a 2 layers map, but Data.MapKeys has 3 field.
}

func ExampleData_init_invalidMapKeys_3() {
	mutex := sync.RWMutex{}
	var m map[int]map[string]int
	d := Data{RWMutex: &mutex, DataPtr: &m, MapKeys: []string{"Student", "Subject"}}
	fmt.Println(d.init(reflect.TypeOf(Score{})))
	// Output:
	// Data.MapKeys[0]: Student, no such field in row struct.
}

func ExampleData_init_invalidMapKeys_4() {
	mutex := sync.RWMutex{}
	var m map[int]map[string]int
	d := Data{RWMutex: &mutex, DataPtr: &m, MapKeys: []string{"Subject", "StudentId"}}
	fmt.Println(d.init(reflect.TypeOf(Score{})))
	// Output:
	// Data.MapKeys[0]: Subject, type string is not assignable to int.
}

func ExampleData_init_invalidValue_1() {
	mutex := sync.RWMutex{}
	var m map[int]map[string]int
	d := Data{
		RWMutex: &mutex,
		DataPtr: &m, MapKeys: []string{"StudentId", "Subject"}, Value: "theScore",
	}
	fmt.Println(d.init(reflect.TypeOf(Score{})))
	// Output:
	// Data.Value: theScore, no such field in row struct.
}

func ExampleData_init_invalidValue_2() {
	mutex := sync.RWMutex{}
	var m map[int]map[string]float32
	d := Data{
		RWMutex: &mutex,
		DataPtr: &m, MapKeys: []string{"StudentId", "Subject"}, Value: "Score",
	}
	fmt.Println(d.init(reflect.TypeOf(Score{})))
	// Output:
	// Data.Value: Score, type int is not assignable to float32.
}

func ExampleData_init_invalidSortedSetUniqueKey_1() {
	mutex := sync.RWMutex{}
	var m map[int]map[string]int
	d := Data{
		RWMutex: &mutex,
		DataPtr: &m, MapKeys: []string{"StudentId", "Subject"}, Value: "Score",
		SortedSetUniqueKey: []string{"Other"},
	}
	fmt.Println(d.init(reflect.TypeOf(Score{})))
	// Output:
	// Data.SortedSetUniqueKey should be empty.
}

func ExampleData_init_invalidSortedSetUniqueKey_2() {
	mutex := sync.RWMutex{}
	var m map[int]map[string][]int
	d := Data{
		RWMutex: &mutex,
		DataPtr: &m, MapKeys: []string{"StudentId", "Subject"}, Value: "Score",
		SortedSetUniqueKey: []string{"Other"},
	}
	fmt.Println(d.init(reflect.TypeOf(Score{})))
	// Output:
	// Data.SortedSetUniqueKey should be empty.
}

func ExampleData_init_invalidSortedSetUniqueKey_3() {
	mutex := sync.RWMutex{}
	var m map[int][]Score
	d := Data{
		RWMutex: &mutex,
		DataPtr: &m, MapKeys: []string{"StudentId"},
		SortedSetUniqueKey: []string{},
	}
	fmt.Println(d.init(reflect.TypeOf(Score{})))
	// Output:
	// Data.SortedSetUniqueKey should not be empty.
}

func ExampleData_init_invalidSortedSetUniqueKey_4() {
	mutex := sync.RWMutex{}
	var m map[int][]Score
	d := Data{
		RWMutex: &mutex,
		DataPtr: &m, MapKeys: []string{"StudentId"},
		SortedSetUniqueKey: []string{"Other"},
	}
	fmt.Println(d.init(reflect.TypeOf(Score{})))
	// Output:
	// Data.SortedSetUniqueKey[0]: Other, no such field in value struct.
}

func ExampleData_init_invalidSortedSetUniqueKey_5() {
	mutex := sync.RWMutex{}
	type score2 struct {
		Score
		ScoreFloat float32
	}
	var m map[int][]score2
	d := Data{
		RWMutex: &mutex,
		DataPtr: &m, MapKeys: []string{"StudentId"},
		SortedSetUniqueKey: []string{"ScoreFloat"},
	}
	fmt.Println(d.init(reflect.TypeOf(score2{})))
	// Output:
	// Data.SortedSetUniqueKey[0]: ScoreFloat, should be a integer or string type.
}

func ExampleData_init_flags1() {
	mutex := sync.RWMutex{}
	var m map[int]map[string][]*int
	d := Data{
		RWMutex: &mutex,
		DataPtr: &m, MapKeys: []string{"StudentId", "Subject"}, Value: "Score",
	}
	fmt.Println(d.init(reflect.TypeOf(Score{})))
	fmt.Println(d.isSortedSets, d.realValueIsPointer)
	// Output:
	// <nil>
	// true true
}

func ExampleData_init_flags2() {
	mutex := sync.RWMutex{}
	var m map[int][]*Score
	d := Data{
		RWMutex: &mutex,
		DataPtr: &m, MapKeys: []string{"StudentId"},
		SortedSetUniqueKey: []string{"Subject"},
	}
	fmt.Println(d.init(reflect.TypeOf(Score{})))
	fmt.Println(d.isSortedSets, d.realValueIsPointer)
	// Output:
	// <nil>
	// true true
}

func ExampleData_init_invalidPrecond_1() {
	mutex := sync.RWMutex{}
	var m map[int]Score
	d := Data{
		RWMutex: &mutex,
		DataPtr: &m, MapKeys: []string{"StudentId"},
		Precond: "None",
	}
	fmt.Println(d.init(reflect.TypeOf(Score{})))
	// Output:
	// Data.Precond: None, no such method for the row struct.
}

func ExampleData_init_invalidPrecond_2() {
	mutex := sync.RWMutex{}
	var m map[int]Score
	d := Data{
		RWMutex: &mutex,
		DataPtr: &m, MapKeys: []string{"StudentId"},
		Precond: "Other",
	}
	fmt.Println(d.init(reflect.TypeOf(Score{})))
	// Output:
	// Data.Precond: Other, should be of "func () bool" form.
}

func ExampleData_init_validPrecond_1() {
	mutex := sync.RWMutex{}
	var m map[int]Score
	d := Data{
		RWMutex: &mutex,
		DataPtr: &m, MapKeys: []string{"StudentId"},
		Precond: "Valid",
	}
	fmt.Println(d.init(reflect.TypeOf(Score{})))
	fmt.Println(d.precondMethodIndex)
	// Output:
	// <nil>
	// 1
}

func ExampleData_init_validPrecond_2() {
	mutex := sync.RWMutex{}
	var m map[int]Score
	d := Data{
		RWMutex: &mutex,
		DataPtr: &m, MapKeys: []string{"StudentId"},
		Precond: "Valid2",
	}
	fmt.Println(d.init(reflect.TypeOf(Score{})))
	fmt.Println(d.precondMethodIndex)
	// Output:
	// <nil>
	// 2
}
