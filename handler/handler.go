package handler

import (
	"encoding/json"
	"log"
	"reflect"
)

type Handler struct {
	rowStruct reflect.Type
	datas     []*Data
	dbQuerier DBQuerier
	loadSql   string
	logger    Logger
}

type DBQuerier interface {
	Query(data interface{}, sql string, args ...interface{}) error
}

type Logger interface {
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
}

func New(
	rowStruct interface{}, datas []Data, db DBQuerier, loadSql string, logger Logger,
) *Handler {
	var rowStrct = reflect.TypeOf(rowStruct)
	if rowStrct.Kind() != reflect.Struct {
		log.Panic("rowStruct is not a struct")
	}
	var initedDatas []*Data
	for i := range datas {
		datas[i].init(rowStrct)
		initedDatas = append(initedDatas, &datas[i])
	}
	if len(initedDatas) == 0 {
		log.Panic("datas is empty")
	}

	return &Handler{
		rowStruct: rowStrct,
		datas:     initedDatas,
		dbQuerier: db,
		loadSql:   loadSql,
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
	var rows = reflect.New(reflect.SliceOf(h.rowStruct)).Elem()
	if err := h.dbQuerier.Query(rows.Addr().Interface(), h.loadSql); err != nil {
		h.logger.Error(err)
		return
	}
	h.Clear()
	h.Save(rows.Interface())
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
