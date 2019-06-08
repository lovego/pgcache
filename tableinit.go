package pgcache

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/lovego/struct_tag"
)

func (t *Table) init() error {
	if t.Name == "" {
		return errors.New("Name should not be empty.")
	}

	t.rowStruct = reflect.TypeOf(t.RowStruct)
	if t.rowStruct.Kind() != reflect.Struct {
		return errors.New("RowStruct is not a struct")
	}

	if t.Columns == "" {
		t.Columns = ColumnsFromStruct(t.rowStruct)
	}

	if t.LoadSql == "" {
		bigColumns := t.BigColumns
		if bigColumns != "" {
			bigColumns = ", " + bigColumns
		}
		t.LoadSql = fmt.Sprintf("SELECT %s %s FROM %s", t.Columns, t.BigColumns, t.Name)
	}

	if len(t.Datas) == 0 {
		return errors.New("Datas should not be empty")
	}
	for i := range t.Datas {
		if err := t.Datas[i].init(t.rowStruct); err != nil {
			return err
		}
	}

	return nil
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
