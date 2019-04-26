package pghandler

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
	fmt.Printf("%s\n%s\n%s\n%s\n", t.Name, t.Columns, t.LoadSql, t.CheckColumns)
	// Output:
	// scores
	// student_id,subject,score
	// select student_id,subject,score from scores
	// student_id,subject,score
}
