package handler

import (
	"fmt"
	"reflect"
	"sync"
)

type Score struct {
	StudentId  int
	Subject    string
	Score      int
	ScoreFloat float32
}

func (s Score) Valid() bool {
	return s.Score >= 0
}

func (s *Score) Valid2() bool {
	return s.Score >= 0
}

func (s *Score) Other() {
}

func ExampleData_init_nilMutex() {
	defer func() {
		fmt.Println(recover())
	}()
	d := Data{}
	d.init(nil)
	// Output:
	// Data.RWMutex is nil.
}

func ExampleData_init_invalidMapPtr_1() {
	defer func() {
		fmt.Println(recover())
	}()
	mutex := sync.RWMutex{}

	d := Data{RWMutex: &mutex, MapPtr: map[int]int{}}
	d.init(nil)
	// Output:
	// Data.Map should be a non nil pointer to a map.
}

func ExampleData_init_invalidMapPtr_2() {
	defer func() {
		fmt.Println(recover())
	}()

	mutex := sync.RWMutex{}
	var p *map[int]int
	d := Data{RWMutex: &mutex, MapPtr: p}
	d.init(nil)
	// Output:
	// Data.Map should be a non nil pointer to a map.
}

func ExampleData_init_invalidMapKeys_1() {
	defer func() {
		fmt.Println(recover())
	}()
	mutex := sync.RWMutex{}
	var m map[int]int
	d := Data{RWMutex: &mutex, MapPtr: &m}
	d.init(nil)
	// Output:
	// Data.Map has depth: 1, but Data.MapKeys has 0 field.
}

func ExampleData_init_invalidMapKeys_2() {
	defer func() {
		fmt.Println(recover())
	}()
	mutex := sync.RWMutex{}
	var m map[int]map[string]int
	d := Data{RWMutex: &mutex, MapPtr: &m, MapKeys: []string{"StudentId", "Subject", "Other"}}
	d.init(reflect.TypeOf(Score{}))
	// Output:
	// Data.Map has depth: 2, but Data.MapKeys has 3 field.
}

func ExampleData_init_invalidMapKeys_3() {
	defer func() {
		fmt.Println(recover())
	}()
	mutex := sync.RWMutex{}
	var m map[int]map[string]int
	d := Data{RWMutex: &mutex, MapPtr: &m, MapKeys: []string{"Student", "Subject"}}
	d.init(reflect.TypeOf(Score{}))
	// Output:
	// Data.MapKeys[0]: Student, no such field in row struct.
}

func ExampleData_init_invalidMapKeys_4() {
	defer func() {
		fmt.Println(recover())
	}()
	mutex := sync.RWMutex{}
	var m map[int]map[string]int
	d := Data{RWMutex: &mutex, MapPtr: &m, MapKeys: []string{"Subject", "StudentId"}}
	d.init(reflect.TypeOf(Score{}))
	// Output:
	// Data.MapKeys[0]: Subject, type string is not assignable to int.
}

func ExampleData_init_invalidMapValue_1() {
	defer func() {
		fmt.Println(recover())
	}()
	mutex := sync.RWMutex{}
	var m map[int]map[string]int
	d := Data{
		RWMutex: &mutex,
		MapPtr:  &m, MapKeys: []string{"StudentId", "Subject"}, MapValue: "theScore",
	}
	d.init(reflect.TypeOf(Score{}))
	// Output:
	// Data.MapValue: theScore, no such field in row struct.
}

func ExampleData_init_invalidMapValue_2() {
	defer func() {
		fmt.Println(recover())
	}()
	mutex := sync.RWMutex{}
	var m map[int]map[string]float32
	d := Data{
		RWMutex: &mutex,
		MapPtr:  &m, MapKeys: []string{"StudentId", "Subject"}, MapValue: "Score",
	}
	d.init(reflect.TypeOf(Score{}))
	// Output:
	// Data.MapValue: Score, type int is not assignable to float32.
}

func ExampleData_init_invalidSortedSetUniqueKey_1() {
	defer func() {
		fmt.Println(recover())
	}()
	mutex := sync.RWMutex{}
	var m map[int]map[string]int
	d := Data{
		RWMutex: &mutex,
		MapPtr:  &m, MapKeys: []string{"StudentId", "Subject"}, MapValue: "Score",
		SortedSetUniqueKey: []string{"Other"},
	}
	d.init(reflect.TypeOf(Score{}))
	// Output:
	// Data.SortedSetUniqueKey should be empty.
}

func ExampleData_init_invalidSortedSetUniqueKey_2() {
	defer func() {
		fmt.Println(recover())
	}()
	mutex := sync.RWMutex{}
	var m map[int]map[string][]int
	d := Data{
		RWMutex: &mutex,
		MapPtr:  &m, MapKeys: []string{"StudentId", "Subject"}, MapValue: "Score",
		SortedSetUniqueKey: []string{"Other"},
	}
	d.init(reflect.TypeOf(Score{}))
	// Output:
	// Data.SortedSetUniqueKey should be empty.
}

func ExampleData_init_invalidSortedSetUniqueKey_3() {
	defer func() {
		fmt.Println(recover())
	}()
	mutex := sync.RWMutex{}
	var m map[int][]Score
	d := Data{
		RWMutex: &mutex,
		MapPtr:  &m, MapKeys: []string{"StudentId"},
		SortedSetUniqueKey: []string{},
	}
	d.init(reflect.TypeOf(Score{}))
	// Output:
	// Data.SortedSetUniqueKey should not be empty.
}

func ExampleData_init_invalidSortedSetUniqueKey_4() {
	defer func() {
		fmt.Println(recover())
	}()
	mutex := sync.RWMutex{}
	var m map[int][]Score
	d := Data{
		RWMutex: &mutex,
		MapPtr:  &m, MapKeys: []string{"StudentId"},
		SortedSetUniqueKey: []string{"Other"},
	}
	d.init(reflect.TypeOf(Score{}))
	// Output:
	// Data.SortedSetUniqueKey[0]: Other, no such field in value struct.
}

func ExampleData_init_invalidSortedSetUniqueKey_5() {
	defer func() {
		fmt.Println(recover())
	}()
	mutex := sync.RWMutex{}
	var m map[int][]Score
	d := Data{
		RWMutex: &mutex,
		MapPtr:  &m, MapKeys: []string{"StudentId"},
		SortedSetUniqueKey: []string{"ScoreFloat"},
	}
	d.init(reflect.TypeOf(Score{}))
	// Output:
	// Data.SortedSetUniqueKey[0]: ScoreFloat, should be a integer or string type.
}

func ExampleData_init_flags1() {
	mutex := sync.RWMutex{}
	var m map[int]map[string][]*int
	d := Data{
		RWMutex: &mutex,
		MapPtr:  &m, MapKeys: []string{"StudentId", "Subject"}, MapValue: "Score",
	}
	d.init(reflect.TypeOf(Score{}))
	fmt.Println(d.isSortedSets, d.realValueIsPointer)
	// Output:
	// true true
}

func ExampleData_init_flags2() {
	mutex := sync.RWMutex{}
	var m map[int][]*Score
	d := Data{
		RWMutex: &mutex,
		MapPtr:  &m, MapKeys: []string{"StudentId"},
		SortedSetUniqueKey: []string{"Subject"},
	}
	d.init(reflect.TypeOf(Score{}))
	fmt.Println(d.isSortedSets, d.realValueIsPointer)
	// Output:
	// true true
}

func ExampleData_init_invalidPrecondMethod_1() {
	defer func() {
		fmt.Println(recover())
	}()
	mutex := sync.RWMutex{}
	var m map[int]Score
	d := Data{
		RWMutex: &mutex,
		MapPtr:  &m, MapKeys: []string{"StudentId"},
		PrecondMethod: "None",
	}
	d.init(reflect.TypeOf(Score{}))
	// Output:
	// Data.PrecondMethod: None, no such method for the row struct.
}

func ExampleData_init_invalidPrecondMethod_2() {
	defer func() {
		fmt.Println(recover())
	}()
	mutex := sync.RWMutex{}
	var m map[int]Score
	d := Data{
		RWMutex: &mutex,
		MapPtr:  &m, MapKeys: []string{"StudentId"},
		PrecondMethod: "Other",
	}
	d.init(reflect.TypeOf(Score{}))
	// Output:
	// Data.PrecondMethod: Other, should be of "func () bool" form.
}

func ExampleData_init_validPrecondMethod_1() {
	mutex := sync.RWMutex{}
	var m map[int]Score
	d := Data{
		RWMutex: &mutex,
		MapPtr:  &m, MapKeys: []string{"StudentId"},
		PrecondMethod: "Valid",
	}
	d.init(reflect.TypeOf(Score{}))
	fmt.Println(d.precondMethodIndex)
	// Output:
	// 1
}

func ExampleData_init_validPrecondMethod_2() {
	mutex := sync.RWMutex{}
	var m map[int]Score
	d := Data{
		RWMutex: &mutex,
		MapPtr:  &m, MapKeys: []string{"StudentId"},
		PrecondMethod: "Valid2",
	}
	d.init(reflect.TypeOf(Score{}))
	fmt.Println(d.precondMethodIndex)
	// Output:
	// 2
}
