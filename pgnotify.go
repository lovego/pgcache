package pgnotify

import (
	"encoding/json"

	"github.com/go-pg/pg"
	"github.com/lovego/errs"
	"github.com/lovego/logger"
)

type Notifier struct {
	DB        *pg.DB
	Listeners map[string]Listener
	Logger    *logger.Logger
}

type Message struct {
	Table  string
	Action string
	Data   json.RawMessage
}

type Listener interface {
	Create(buf []byte)
	Update(buf []byte)
	Delete(buf []byte)
}

func (n *Notifier) Start() error {
	if err := n.CreateFunction(); err != nil {
		return err
	}
	if err := n.CreateTriggers(); err != nil {
		return err
	}
	go n.Listen()
	return nil
}

func (n *Notifier) Listen() {
	ch := n.DB.Listen("pgnotify").Channel()
	for notice := range ch {
		var msg Message
		if err := json.Unmarshal([]byte(notice.Payload), &msg); err != nil {
			n.Logger.Error(err)
		}
		if listener := n.Listeners[msg.Table]; listener != nil {
			switch msg.Action {
			case "INSERT":
				listener.Create(msg.Data)
			case "UPDATE":
				listener.Update(msg.Data)
			case "DELETE":
				listener.Delete(msg.Data)
			default:
				n.Logger.Errorf("unexpected msg: %+v", msg)
			}
		} else {
			n.Logger.Errorf("unexpected msg: %+v", msg)
		}
	}
}

func (n *Notifier) CreateFunction() error {
	if _, err := n.DB.Exec(`CREATE OR REPLACE FUNCTION pgnotify() RETURNS TRIGGER AS $$
	DECLARE
		data json;
		notification json;
	BEGIN
		IF (TG_OP = 'DELETE') THEN
				data = row_to_json(OLD);
		ELSE
				data = row_to_json(NEW);
		END IF;
		notification = json_build_object(
			'table', TG_TABLE_NAME, 'action', TG_OP, 'data', data
		);
		PERFORM pg_notify('pgnotify', notification::text);
		RETURN NULL;
	END;
$$ LANGUAGE plpgsql;`); err != nil {
		return errs.Trace(err)
	}
	return nil
}

func (n *Notifier) CreateTriggers() error {
	for tableName := range n.Listeners {
		var count int
		if _, err := n.DB.Query(pg.Scan(&count),
			`select count(*) as count from pg_trigger
			where tgrelid = ?::regclass and tgname ='?_pgnotify' and not tgisinternal`,
			tableName, pg.Q(tableName),
		); err != nil {
			return errs.Trace(err)
		}
		if count <= 0 {
			if _, err := n.DB.Exec(`CREATE TRIGGER ?_pgnotify
				AFTER INSERT OR UPDATE OR DELETE ON ?
				FOR EACH ROW EXECUTE PROCEDURE pgnotify()`,
				pg.Q(tableName), pg.Q(tableName),
			); err != nil {
				return errs.Trace(err)
			}
		}
	}
	return nil
}
