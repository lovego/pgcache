package pgcache

import (
	"fmt"
)

func ExampleTable_init() {
	t := Table{
		Name: "scores",
		RowStruct: struct {
			Score
			Z bool `json:"-"`
		}{},
	}
	t.init(testQuerier{}, testLogger)
	fmt.Println(t.Columns)
	fmt.Println(t.BigColumns)
	fmt.Println(t.LoadSql)
	// Output:
	// student_id,subject,score
	//
	// SELECT student_id,subject,score  FROM scores
}

func ExampleTable_init_bigColumns() {
	t := Table{
		Name:               "scores",
		RowStruct:          Score{},
		BigColumns:         "score",
		BigColumnsLoadKeys: []string{"StudentId", "Subject"},
	}
	t.init(testQuerier{}, testLogger)
	fmt.Println(t.Columns)
	fmt.Println(t.BigColumns)
	fmt.Println(t.bigColumnsLoadSql)
	fmt.Println(t.LoadSql)

	// Output:
	// student_id,subject
	// score
	// SELECT score FROM scores WHERE student_id = %s AND subject = %s
	// SELECT student_id,subject ,score FROM scores
}
