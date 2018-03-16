package pgnotify

import (
	"encoding/json"
	"strings"

	"github.com/go-pg/pg"
	"github.com/lovego/logger"
)

type Notifier struct {
	db       *pg.DB
	logger   *logger.Logger
	listener *pg.Listener
	handlers map[string]Handler
}

type Handler interface {
	Create(table string, buf []byte)
	Update(table string, buf []byte)
	Delete(table string, buf []byte)
}

type Message struct {
	Action string
	Data   json.RawMessage
}

func New(db *pg.DB, logger *logger.Logger) (*Notifier, error) {
	if err := CreateFunction(db); err != nil {
		return nil, err
	}
	return &Notifier{
		db: db, logger: logger, handlers: make(map[string]Handler),
	}, nil
}

func (n *Notifier) Notify(table string, handler Handler) error {
	n.handlers[table] = handler
	if err := CreateTriggerIfNotExists(n.db, table); err != nil {
		return err
	}
	if n.listener == nil {
		n.listener = n.db.Listen("pgnotify_" + table)
		go n.listen()
	} else {
		n.listener.Listen("pgnotify_" + table)
	}
	return nil
}

func (n *Notifier) listen() {
	var channel = n.listener.Channel()
	for notice := range channel {
		var msg Message
		if err := json.Unmarshal([]byte(notice.Payload), &msg); err != nil {
			n.logger.Error(err)
		}
		var table = strings.TrimPrefix(notice.Channel, "pgnotify_")
		if handler := n.handlers[table]; handler != nil {
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
		} else {
			n.logger.Errorf("unexpected msg: %+v", msg)
		}
	}
}
