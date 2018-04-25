/*
适用于低频修改且量小的数据，单线程操作。
*/
package pgnotify

import (
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/lovego/errs"
	"github.com/lovego/logger"
)

type Handler interface {
	Create(table string, buf []byte)
	Update(table string, buf []byte)
	Delete(table string, buf []byte)
	ConnLoss(table string)
}

type Notifier struct {
	db       *sql.DB
	listener *pq.Listener
	logger   *logger.Logger
	handlers map[string]Handler
}

type Message struct {
	Action string
	Data   json.RawMessage
}

func New(dbAddr string, logger *logger.Logger) (*Notifier, error) {
	db, err := sql.Open(`postgres`, dbAddr)
	if err != nil {
		return nil, errs.Trace(err)
	}
	if err := db.Ping(); err != nil {
		return nil, errs.Trace(err)
	}
	db.SetConnMaxLifetime(time.Minute)
	db.SetMaxIdleConns(1)
	db.SetMaxOpenConns(1)

	if err := CreateFunction(db); err != nil {
		return nil, err
	}
	n := &Notifier{
		db: db, logger: logger, handlers: make(map[string]Handler),
	}
	n.listener = pq.NewListener(dbAddr, time.Nanosecond, time.Minute, n.eventLogger)
	go n.listen()
	return n, nil
}

func (n *Notifier) Notify(table string, handler Handler) error {
	if err := CreateTriggerIfNotExists(n.db, table); err != nil {
		return err
	}
	n.handlers[table] = handler
	channel := "pgnotify_" + table
	if err := n.listener.Listen(channel); err != nil {
		return errs.Trace(err)
	}
	n.listener.Notify <- &pq.Notification{Channel: channel, Extra: "reload"}
	return nil
}

func (n *Notifier) listen() {
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
		return
	}

	var msg Message
	if err := json.Unmarshal([]byte(notice.Extra), &msg); err != nil {
		n.logger.Error(err)
	}
	switch msg.Action {
	case "INSERT":
		handler.Create(table, msg.Data)
	case "UPDATE":
		handler.Update(table, msg.Data)
	case "DELETE":
		handler.Delete(table, msg.Data)
	default:
		n.logger.Errorf("unexpected msg: %+v", msg)
	}
}

func (n *Notifier) eventLogger(event pq.ListenerEventType, err error) {
	if err != nil {
		n.logger.Error(event, err)
	}
}
