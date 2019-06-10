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
	fmt.Printf("%s\n%s\n%s\n%s\n", t.Name, t.Columns, t.BigColumns, t.LoadSql)
	// Output:
	// scores
	// student_id,subject,score
	//
	// SELECT student_id,subject,score  FROM scores
}
