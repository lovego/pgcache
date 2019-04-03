/*
适用于低频修改且量小的数据，单线程操作。
*/
package pgnotify

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

type Notifier struct {
	db       *sql.DB // db to create func and triggers
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

func New(dbAddr string, logger Logger) (*Notifier, error) {
	db, err := getDb(dbAddr)
	if err != nil {
		return nil, err
	}
	if err := createPGFunction(db); err != nil {
		return nil, err
	}
	n := &Notifier{
		db:       db,
		logger:   logger,
		handlers: make(map[string]Handler),
		inited:   make(map[string]chan struct{}),
	}
	n.listener = pq.NewListener(dbAddr, time.Second, time.Minute, n.eventLogger)
	go n.loop()
	return n, nil
}

func (n *Notifier) Notify(handler TableHandler) error {
	return n.Listen(handler.TableName(), handler.Columns(), handler.CheckColumns(), handler)
}

// Listen a table, and send "columns" to the handler when a row is created/updated/deleted.
// When a row is Updated, only if some "checkColumns" has changed, the "columns" will be send to
// the handler.
func (n *Notifier) Listen(table string, columns, checkColumns string, handler Handler) error {
	if strings.IndexByte(table, '.') < 0 {
		table = "public." + table
	}
	if _, ok := n.handlers[table]; ok {
		return fmt.Errorf("pgnotify: table '%s' is aready listened.", table)
	}
	if err := createTrigger(n.db, table, columns, checkColumns); err != nil {
		return err
	}
	n.handlers[table] = handler
	n.inited[table] = make(chan struct{})
	channel := "pgnotify_" + table
	if err := n.listener.Listen(channel); err != nil {
		return errs.Trace(err)
	}
	n.listener.Notify <- &pq.Notification{Channel: channel, Extra: "reload"}
	<-n.inited[table]
	return nil
}

func (n *Notifier) Unlisten(table string) error {
	if strings.IndexByte(table, '.') < 0 {
		table = "public." + table
	}
	channel := "pgnotify_" + table
	if err := n.listener.Unlisten(channel); err != nil {
		return errs.Trace(err)
	}
	return nil
}

func (n *Notifier) UnlistenAll() error {
	if err := n.listener.UnlistenAll(); err != nil {
		return errs.Trace(err)
	}
	return nil
}

func (n *Notifier) loop() {
	for {
		select {
		case notice := <-n.listener.Notify:
			n.handle(notice)
		case <-time.After(time.Minute):
			go n.listener.Ping()
		}
	}
}

func (n *Notifier) handle(notice *pq.Notification) {
	if notice == nil { // connection loss
		for table, handler := range n.handlers {
			handler.ConnLoss(table)
		}
		return
	}

	var table = strings.TrimPrefix(notice.Channel, "pgnotify_")
	handler := n.handlers[table]
	if handler == nil {
		n.logger.Errorf("unexpected notification: %+v", notice)
	}
	if notice.Extra == "reload" {
		handler.ConnLoss(table)
		n.inited[table] <- struct{}{}
		close(n.inited[table])
		return
	}

	var msg message
	if err := json.Unmarshal([]byte(notice.Extra), &msg); err != nil {
		n.logger.Error(err)
	}
	switch msg.Action {
	case "INSERT":
		handler.Create(table, msg.New)
	case "UPDATE":
		handler.Update(table, msg.Old, msg.New)
	case "DELETE":
		handler.Delete(table, msg.Old)
	default:
		n.logger.Errorf("unexpected msg: %+v", msg)
	}
}

func (n *Notifier) eventLogger(event pq.ListenerEventType, err error) {
	if err != nil {
		n.logger.Error(event, err)
	}
}

func (n *Notifier) DB() *sql.DB {
	return n.db
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
	// don't keep idle connections. only used to create func and triggers.
	db.SetMaxIdleConns(0)
	db.SetMaxOpenConns(1)

	return db, nil
}
