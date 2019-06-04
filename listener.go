/*
适用于低频修改且量小的数据，单线程操作。
*/
package pglistener

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/lovego/errs"
	"github.com/lovego/pglistener/cache/manage"
)

type Listener struct {
	db       *sql.DB // db to create func and triggers
	dbName   string
	listener *pq.Listener
	logger   Logger
	handlers map[string]Handler
	inited   map[string]chan struct{}
}

type Handler interface {
	Create(table string, content []byte)
	Update(table string, oldContent, newContent []byte)
	Delete(table string, content []byte)
	ConnLoss(table string)
}

type TableHandler interface {
	Handler
	TableName() string
	Columns() string
	CheckColumns() string
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

func New(dbAddr string, logger Logger) (*Listener, error) {
	var dbName string
	if u, err := url.Parse(dbAddr); err != nil {
		return nil, err
	} else {
		dbName = strings.TrimPrefix(u.Path, "/")
	}

	db, err := getDb(dbAddr)
	if err != nil {
		return nil, err
	}
	if err := createPGFunction(db); err != nil {
		return nil, err
	}
	l := &Listener{
		db:       db,
		dbName:   dbName,
		logger:   logger,
		handlers: make(map[string]Handler),
		inited:   make(map[string]chan struct{}),
	}
	l.listener = pq.NewListener(dbAddr, time.Second, time.Minute, l.eventLogger)
	go l.loop()
	return l, nil
}

func (l *Listener) ListenTable(handler TableHandler) error {
	return l.Listen(handler.TableName(), handler.Columns(), handler.CheckColumns(), handler)
}

// Listen a table, and send "columns" to the handler when a row is created/updated/deleted.
// When a row is Updated, only if some "checkColumns" has changed, the "columns" will be send to
// the handler.
func (l *Listener) Listen(table string, columns, checkColumns string, handler Handler) error {
	manage.TryRegister(l.dbName, table, handler)

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
	l.listener.Notify <- &pq.Notification{Channel: l.GetChannel(table), Extra: "reload"}
	<-l.inited[table]
	return nil
}

func (l *Listener) Unlisten(table string) error {
	manage.Unregister(l.dbName, table)

	if strings.IndexByte(table, '.') < 0 {
		table = "public." + table
	}
	if err := l.listener.Unlisten(l.GetChannel(table)); err != nil {
		return errs.Trace(err)
	}
	return nil
}

func (l *Listener) UnlistenAll() error {
	manage.UnregisterDB(l.dbName)

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
	if notice.Extra == "reload" {
		handler.ConnLoss(table)
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
