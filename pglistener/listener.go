/*
适用于低频修改且量小的数据，单线程操作。
*/
package pglistener

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/lovego/errs"
)

// Listen for INSERT/UPDATE/DELETE events of postgresql's table,
// and pass the events to defined handlers.
type Listener struct {
	db       *sql.DB // db to create func and triggers
	listener *pq.Listener
	logger   Logger
	handlers map[string]Handler
	inited   map[string]chan struct{}
}

type Handler interface {
	Init(table string)
	Create(table string, content []byte)
	Update(table string, oldContent, newContent []byte)
	Delete(table string, content []byte)
	ConnLoss(table string)
}

type Logger interface {
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
}

type message struct {
	Action string
	Old    json.RawMessage
	New    json.RawMessage
}

func New(dbAddr string, db *sql.DB, logger Logger) (*Listener, error) {
	if db == nil {
		var err error
		if db, err = getDb(dbAddr); err != nil {
			return nil, err
		}
	}
	if err := createPGFunction(db); err != nil {
		return nil, err
	}
	l := &Listener{
		db:       db,
		logger:   logger,
		handlers: make(map[string]Handler),
		inited:   make(map[string]chan struct{}),
	}
	l.listener = pq.NewListener(dbAddr, time.Second, time.Minute, l.eventLogger)
	go l.loop()
	return l, nil
}

// Listen a table and notify the handler with "columns" when a row is created or updated or deleted.
// When a row is updated, the handler is notified only if some "columns" or "checkColumns" has changed.
func (l *Listener) Listen(table string, columns, checkColumns string, handler Handler) error {
	if strings.IndexByte(table, '.') < 0 {
		table = "public." + table
	}
	if _, ok := l.handlers[table]; ok {
		return fmt.Errorf("pglistener: table '%s' is aready listened.", table)
	}
	if err := createTrigger(l.db, table, columns, checkColumns); err != nil {
		return err
	}
	l.handlers[table] = handler
	l.inited[table] = make(chan struct{})
	if err := l.listener.Listen(l.GetChannel(table)); err != nil {
		return errs.Trace(err)
	}
	l.listener.Notify <- &pq.Notification{Channel: l.GetChannel(table), Extra: "init"}
	<-l.inited[table]
	return nil
}

func (l *Listener) Unlisten(table string) error {
	if strings.IndexByte(table, '.') < 0 {
		table = "public." + table
	}
	if err := l.listener.Unlisten(l.GetChannel(table)); err != nil {
		return errs.Trace(err)
	}
	return nil
}

func (l *Listener) UnlistenAll() error {
	if err := l.listener.UnlistenAll(); err != nil {
		return errs.Trace(err)
	}
	return nil
}

func (l *Listener) loop() {
	for {
		select {
		case notice := <-l.listener.Notify:
			l.handle(notice)
		case <-time.After(time.Minute):
			go l.listener.Ping()
		}
	}
}

func (l *Listener) handle(notice *pq.Notification) {
	if notice == nil { // connection loss
		for table, handler := range l.handlers {
			handler.ConnLoss(table)
		}
		return
	}

	var table = l.GetTable(notice.Channel)
	handler := l.handlers[table]
	if handler == nil {
		l.logger.Errorf("unexpected Notification: %+v", notice)
	}
	if notice.Extra == "init" {
		handler.Init(table)
		l.inited[table] <- struct{}{}
		close(l.inited[table])
		return
	}

	var msg message
	if err := json.Unmarshal([]byte(notice.Extra), &msg); err != nil {
		l.logger.Error(err)
	}
	switch msg.Action {
	case "INSERT":
		handler.Create(table, msg.New)
	case "UPDATE":
		handler.Update(table, msg.Old, msg.New)
	case "DELETE":
		handler.Delete(table, msg.Old)
	default:
		l.logger.Errorf("unexpected msg: %+v", msg)
	}
}

func (l *Listener) GetChannel(table string) string {
	return "pgnotify_" + table
}

func (l *Listener) GetTable(channel string) string {
	return strings.TrimPrefix(channel, "pgnotify_")
}

func (l *Listener) eventLogger(event pq.ListenerEventType, err error) {
	if err != nil {
		l.logger.Error(event, err)
	}
}

func (l *Listener) DB() *sql.DB {
	return l.db
}

func getDb(dbAddr string) (*sql.DB, error) {
	db, err := sql.Open(`postgres`, dbAddr)
	if err != nil {
		return nil, errs.Trace(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, errs.Trace(err)
	}
	db.SetConnMaxLifetime(time.Minute)
	db.SetMaxIdleConns(1)
	db.SetMaxOpenConns(1)

	return db, nil
}
