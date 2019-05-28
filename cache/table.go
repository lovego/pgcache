package cache

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/lovego/struct_tag"
)

type Table struct {
	// The name of the table to listen, required.
	Name string
	// The columns of the table to listen.
	// If empty, the fields of rowStruct is used. The field name is converted to underscore style,
	// and field with `json:"-"` tags is ignored.
	Columns string
	// when a row is updated, if none the "CheckColumns" has changed, the event is dropped.
	// If empty, the fields of rowStruct is used. The field name is converted to underscore style,
	// and field with `json:"-"` tags is ignored.
	CheckColumns string
	// The sql to load data when a handler is inited or the db connection losted.
	// If empty, the fields of rowStruct is used to make a SELECT sql FROM "NAME".
	// The field name is converted to underscore style, and field with `json:"-"` tags is ignored.
	LoadSql string
}

func (t *Table) init(rowStruct reflect.Type) {
	if t.Name == "" && t.LoadSql == "" {
		log.Panic("both Name and LoadSql are empty.")
	}
	if t.Columns == "" {
		t.Columns = ColumnsFromStruct(rowStruct)
	}
	if t.CheckColumns == "" {
		t.CheckColumns = ColumnsFromStruct(rowStruct)
	}
	if t.LoadSql == "" {
		t.LoadSql = fmt.Sprintf(
			"SELECT %s FROM %s", ColumnsFromStruct(rowStruct), t.Name,
		)
	}
}

func ColumnsFromStruct(rowStruct reflect.Type) string {
	var result []string
	traverseStructFields(rowStruct, func(field reflect.StructField) {
		result = append(result, Field2Column(field.Name))
	})
	return strings.Join(result, ",")
}

func traverseStructFields(typ reflect.Type, fn func(field reflect.StructField)) bool {
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return false
	}
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if (!field.Anonymous || !traverseStructFields(field.Type, fn)) &&
			(field.Name[0] >= 'A' && field.Name[0] <= 'Z') {
			if value, ok := struct_tag.Lookup(string(field.Tag), `json`); !ok || value != "-" {
				fn(field)
			}
		}
	}
	return true
}

/* 单词边界有两种
1. 非大写字符，且下一个是大写字符
2. 大写字符，且下一个是大写字符，且下下一个是非大写字符
*/
func Field2Column(str string) string {
	var slice []string
	start := 0
	for end, char := range str {
		if end+1 < len(str) {
			next := str[end+1]
			if char < 'A' || char > 'Z' {
				if next >= 'A' && next <= 'Z' { // 非大写下一个是大写
					slice = append(slice, str[start:end+1])
					start, end = end+1, end+1
				}
			} else if end+2 < len(str) && (next >= 'A' && next <= 'Z') {
				if next2 := str[end+2]; next2 < 'A' || next2 > 'Z' {
					slice = append(slice, str[start:end+1])
					start, end = end+1, end+1
				}
			}
		} else {
			slice = append(slice, str[start:end+1])
		}
	}
	return strings.ToLower(strings.Join(slice, "_"))
}
