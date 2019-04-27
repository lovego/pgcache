package cache

import (
	"encoding/json"
	"log"
	"reflect"
)

// A Handler to cache table data.
type Handler struct {
	table     Table
	rowStruct reflect.Type
	datas     []*Data
	dbQuerier DBQuerier
	logger    Logger
}

type DBQuerier interface {
	Query(data interface{}, sql string, args ...interface{}) error
}

type Logger interface {
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
}

// New return a cache handler.
// Param table is the table to cache.
// Param rowStruct is the struct to receive a table row.
// Param datas is the maps to store all table rows.
// Param db and logger are interfaces to do db query and error logging.
func New(
	table Table, rowStruct interface{}, datas []Data, db DBQuerier, logger Logger,
) *Handler {
	var rowStrct = reflect.TypeOf(rowStruct)
	if rowStrct.Kind() != reflect.Struct {
		log.Panic("rowStruct is not a struct")
	}
	table.init(rowStrct)
	var initedDatas []*Data
	for i := range datas {
		datas[i].init(rowStrct)
		initedDatas = append(initedDatas, &datas[i])
	}
	if len(initedDatas) == 0 {
		log.Panic("datas is empty")
	}

	return &Handler{
		table:     table,
		rowStruct: rowStrct,
		datas:     initedDatas,
		dbQuerier: db,
		logger:    logger,
	}
}

func (h *Handler) Create(table string, content []byte) {
	h.save(content)
}

func (h *Handler) Update(table string, oldContent, newContent []byte) {
	h.remove(oldContent)
	h.save(newContent)
}

func (h *Handler) Delete(table string, content []byte) {
	h.remove(content)
}

func (h *Handler) ConnLoss(table string) {
	if err := h.Reload(); err != nil {
		h.logger.Error(err)
	}
}

func (h *Handler) Reload() error {
	var rows = reflect.New(reflect.SliceOf(h.rowStruct)).Elem()
	if err := h.dbQuerier.Query(rows.Addr().Interface(), h.table.LoadSql); err != nil {
		return err
	}
	h.Clear()
	h.Save(rows.Interface())
	return nil
}

func (h *Handler) save(content []byte) {
	var row = reflect.New(h.rowStruct).Elem()
	if err := json.Unmarshal(content, row.Addr().Interface()); err != nil {
		h.logger.Error(err)
		return
	}
	for _, d := range h.datas {
		d.save(row)
	}
}

func (h *Handler) remove(content []byte) {
	var row = reflect.New(h.rowStruct).Elem()
	if err := json.Unmarshal(content, row.Addr().Interface()); err != nil {
		h.logger.Error(err)
		return
	}
	for _, d := range h.datas {
		d.remove(row)
	}
}

func (h *Handler) Clear() {
	for _, d := range h.datas {
		d.clear()
	}
}

func (h *Handler) Save(rows interface{}) {
	rowsV := reflect.ValueOf(rows)
	for i := 0; i < rowsV.Len(); i++ {
		row := rowsV.Index(i)
		for _, d := range h.datas {
			d.save(row)
		}
	}
}

func (h *Handler) Remove(rows interface{}) {
	rowsV := reflect.ValueOf(rows)
	for i := 0; i < rowsV.Len(); i++ {
		row := rowsV.Index(i)
		for _, d := range h.datas {
			d.remove(row)
		}
	}
}

func (h *Handler) TableName() string {
	return h.table.Name
}

func (h *Handler) Columns() string {
	return h.table.Columns
}

func (h *Handler) CheckColumns() string {
	return h.table.CheckColumns
}
