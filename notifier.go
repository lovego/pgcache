/*
适用于低频修改且量小的数据，单线程操作。
*/
package pgnotify

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/lovego/errs"
	"github.com/lovego/logger"
)

type Handler interface {
	Create(table string, buf []byte)
	Update(table string, oldBuf, newBuf []byte)
	Delete(table string, buf []byte)
	ConnLoss(table string)
}

type Notifier struct {
	db       *sql.DB
	listener *pq.Listener
	logger   *logger.Logger
	handlers map[string]Handler
	inited   map[string]chan struct{}
}

type message struct {
	Action string
	Old    json.RawMessage
	New    json.RawMessage
}

func New(dbAddr string, logger *logger.Logger) (*Notifier, error) {
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
	if os.Getenv("GOENV") == "production" {
		db.SetMaxIdleConns(1)
	} else {
		db.SetMaxIdleConns(0)
	}
	db.SetMaxOpenConns(1)

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
	go n.listen()
	return n, nil
}

func (n *Notifier) DB() *sql.DB {
	return n.db
}

func (n *Notifier) Notify(table string, columnsToNotify, columnsToCheck string, handler Handler) error {
	if _, ok := n.handlers[table]; ok {
		return fmt.Errorf("pgnotify: the trigger of table '%s' aready exists.", table)
	}
	if err := createTrigger(n.db, table, columnsToNotify, columnsToCheck); err != nil {
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
