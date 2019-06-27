package pgcache

import (
	"database/sql"
	"net/url"
	"strings"

	"github.com/lovego/pgcache/manage"
	"github.com/lovego/pgcache/pglistener"
)

type DB struct {
	name      string
	listener  *pglistener.Listener
	dbQuerier DBQuerier
	logger    Logger
}

type DBQuerier interface {
	Query(data interface{}, sql string, args ...interface{}) error
	GetDB() *sql.DB
}

type Logger interface {
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
}

func New(dbAddr string, dbQuerier DBQuerier, logger Logger) (*DB, error) {
	var dbName string
	if uri, err := url.Parse(dbAddr); err != nil {
		return nil, err
	} else {
		dbName = strings.TrimPrefix(uri.Path, "/")
	}
	listener, err := pglistener.New(dbAddr, dbQuerier.GetDB(), logger)
	if err != nil {
		return nil, err
	}
	return &DB{name: dbName, listener: listener, dbQuerier: dbQuerier, logger: logger}, nil
}

func (db *DB) Add(table *Table) (*Table, error) {
	if err := table.init(db.name, db.dbQuerier, db.logger); err != nil {
		return nil, err
	}
	if err := db.listener.Listen(table.Name, table.Columns, table.BigColumns, table); err != nil {
		return nil, err
	}
	if err := manage.Register(db.name, table.Name, table); err != nil {
		return nil, err
	}
	return table, nil
}

func (db *DB) Remove(table string) error {
	manage.Unregister(db.name, table)
	return db.listener.Unlisten(table)
}

func (db *DB) RemoveAll() error {
	manage.UnregisterDB(db.name)
	return db.listener.UnlistenAll()
}
