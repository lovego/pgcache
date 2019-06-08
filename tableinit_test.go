package pgcache

import (
	"fmt"
	"reflect"
)

func ExampleTable_init() {
	t := Table{Name: "scores"}
	t.init(reflect.TypeOf(struct {
		Score
		Z bool `json:"-"`
	}{}))
	fmt.Printf("%s\n%s\n%s\n%s\n", t.Name, t.Columns, t.CheckColumns, t.LoadSql)
	// Output:
	// scores
	// student_id,subject,score
	// student_id,subject,score
	// SELECT student_id,subject,score FROM scores
}
