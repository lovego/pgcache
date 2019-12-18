package pgcache

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/lovego/bsql"
	"github.com/lovego/bsql/scan"
	"github.com/lovego/pgcache/manage"
)

// A Handler to cache table data.
type Table struct {
	dbName string

	// The name of the table to cache, required.
	Name string

	// The struct to receive a table row.
	RowStruct interface{}

	// The columns of the table to cache. It's got from the pg_notify payload, it must be less than
	// 8000 bytes, use "BigColumns" if necessarry.
	// If empty, the fields of "RowStruct" which is not "BigColumns" are used.
	// The field name is converted to underscore style, and field with `json:"-"` tag is ignored.
	Columns string

	// The big columns of the table to cache. It's got by a seperate query.
	BigColumns string
	// The unique fields to load "BigColumns" from db. If empty, and "RowStruct" has a "Id" Field,
	// it's used as "BigColumnsLoadKeys".
	BigColumnsLoadKeys []string
	// sql to load "BigColumns"
	bigColumnsLoadSql string

	// The sql used to load initial data when a table is cached, or reload table data when the db
	// connection lost. If empty, "Columns" and "BigColumns" is used to make a SELECT sql FROM "NAME".
	LoadSql string

	// Datas is the maps to store table rows.
	Datas []*Data

	// db querier to load data from a table.
	dbQuerier DBQuerier

	// errors are logged using this logger.
	logger Logger

	rowStruct reflect.Type
}

func (t *Table) Init(table string) {
	if err := t.Reload(); err != nil {
		t.Error(err)
	}
}

func (t *Table) Create(table string, content []byte) {
	t.save(content)
}

func (t *Table) Update(table string, oldContent, newContent []byte) {
	t.remove(oldContent)
	t.save(newContent)
}

func (t *Table) Delete(table string, content []byte) {
	t.remove(content)
}

func (t *Table) ConnLoss(table string) {
	if err := t.Reload(); err != nil {
		t.Error("connection loss: " + err.Error())
	} else {
		t.Error("connection loss")
	}
}

func (t *Table) Reload() error {
	var rows = reflect.New(reflect.SliceOf(t.rowStruct)).Elem()
	start := time.Now()
	err := t.dbQuerier.Query(rows.Addr().Interface(), t.LoadSql)
	msg := fmt.Sprintf("pgcache reload queryTime: %6v, ", time.Since(start).Round(time.Millisecond))
	if err != nil {
		log.Printf("%s \t%s.%s\n", msg, t.dbName, t.Name)
		return fmt.Errorf("reload: %v", err)
	}
	t.Clear()
	t.Save(rows.Interface())
	log.Printf("%s fullTime: %6v, \t%s.%s\n", msg, time.Since(start).Round(time.Millisecond),
		t.dbName, t.Name)
	return nil
}

func (t *Table) Clear() {
	for _, d := range t.Datas {
		d.clear()
	}
}

func (t *Table) Save(rows interface{}) {
	rowsV := reflect.ValueOf(rows)
	for i := 0; i < rowsV.Len(); i++ {
		row := rowsV.Index(i)
		for _, d := range t.Datas {
			d.save(row)
		}
	}
}

func (t *Table) Remove(rows interface{}) {
	rowsV := reflect.ValueOf(rows)
	for i := 0; i < rowsV.Len(); i++ {
		row := rowsV.Index(i)
		for _, d := range t.Datas {
			d.remove(row)
		}
	}
}

func (t *Table) GetDatas() []manage.Data {
	result := make([]manage.Data, len(t.Datas))
	for i, data := range t.Datas {
		result[i] = data
	}
	return result
}

func (t *Table) save(content []byte) {
	var row = reflect.New(t.rowStruct).Elem()
	if err := jsonUnmarshal(content, row); err != nil {
		t.Error(err)
		return
	}
	if t.BigColumns != "" {
		var params = make([]interface{}, len(t.BigColumnsLoadKeys))
		for i, key := range t.BigColumnsLoadKeys {
			params[i] = bsql.V(row.FieldByName(key).Interface())
		}
		if err := t.dbQuerier.Query(row.Addr().Interface(), fmt.Sprintf(
			t.bigColumnsLoadSql, params...,
		)); err != nil {
			t.Error(err)
			return
		}
	}
	for _, d := range t.Datas {
		d.save(row)
	}
}

func (t *Table) remove(content []byte) {
	var row = reflect.New(t.rowStruct).Elem()
	if err := jsonUnmarshal(content, row); err != nil {
		t.Error(err)
		return
	}
	for _, d := range t.Datas {
		d.remove(row)
	}
}

func (t *Table) Error(err interface{}) {
	t.logger.Errorf("pgcache (%s.%s) %v", t.dbName, t.Name, err)
}

func jsonUnmarshal(content []byte, row reflect.Value) error {
	var m = map[string]json.RawMessage{}

	if err := json.Unmarshal(content, &m); err != nil {
		return err
	}

	for k, v := range m {
		if field := row.FieldByName(scan.Column2Field(k)); field.IsValid() {
			if err := json.Unmarshal(v, field.Addr().Interface()); err != nil {
				return err
			}
		}
	}

	return nil
}
