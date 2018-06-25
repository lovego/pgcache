/*
测试命令:
PG_DATA_SOURCE="postgres://USERNAME:PASSWORD@HOSTNAME:PORT/DBNAME?sslmode=disable" go test
*/
package pgnotify

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"encoding/json"

	"fmt"

	"bytes"

	"github.com/lovego/errs"
	"github.com/lovego/logger"
)

func TestNotifier(t *testing.T) {
	var addr = constructDataSource()
	db, err := sql.Open(`postgres`, addr)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	notifier, err := New(addr, logger.New("", os.Stderr, nil))
	if err != nil {
		t.Fatal(errs.WithStack(err))
	}

	testNotifyActions(notifier, db, t)
	testNotifiedColumnsAsExpected(notifier, db, t)
}

func constructDataSource() string {
	dataSource := "postgres://travis:123456@localhost:5433/travis?sslmode=disable"

	if ds, ok := os.LookupEnv("PG_DATA_SOURCE"); ok {
		dataSource = ds
	}

	return dataSource
}

func testNotifyActions(notifier *Notifier, db *sql.DB, t *testing.T) {
	table := "pgnotify_actions"
	createTable(db, table, t)

	h := triggerActionsHandler{t: t}
	startPGNotify(notifier, table, []string{"time", "id", "name"}, &h, t)

	_, err := db.Exec(fmt.Sprintf(`INSERT INTO %s(name, time) VALUES ('李雷', now())`, table))
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(fmt.Sprintf(`UPDATE %s SET NAME = '韩梅梅'`, table))
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(fmt.Sprintf(`DELETE FROM %s`, table))
	if err != nil {
		t.Fatal(err)
	}

	ensureNotifyReached()

	if h.connLoss != 1 || h.create != 1 || h.update != 1 && h.delete != 1 {
		t.Errorf("expected: %+v", h)
	}
}

func testNotifiedColumnsAsExpected(notifier *Notifier, db *sql.DB, t *testing.T) {
	tablePrefix := "pgnotify_columns"

	testCreate(notifier, db, tablePrefix, t)
	testUpdate(notifier, db, tablePrefix, t)
	testDelete(notifier, db, tablePrefix, t)
}

func testCreate(notifier *Notifier, db *sql.DB, tablePrefix string, t *testing.T) {
	testColumns(notifier, db, tablePrefix+"_all", []string{"id", "name", "time"}, nil, t)
	testColumns(notifier, db, tablePrefix+"_some", []string{"id", "name"}, []string{"id", "name"}, t)
}

func testColumns(
	notifier *Notifier,
	db *sql.DB,
	table string,
	expectedColumns []string,
	actualColumns []string,
	t *testing.T,
) {
	createTable(db, table, t)

	h := notifiedColumnsHandler{}

	startPGNotify(notifier, table, actualColumns, &h, t)

	_, err := db.Exec(fmt.Sprintf(`INSERT INTO %s(name, time) VALUES ('李雷', now())`, table))
	if err != nil {
		t.Fatal(err)
	}

	ensureNotifyReached()

	testColumnsAsExpected(expectedColumns, h.pgNEW, t)
}

func testUpdate(notifier *Notifier, db *sql.DB, tablePrefix string, t *testing.T) {
	table := tablePrefix + "_update"
	createTable(db, table, t)

	_, err := db.Exec(fmt.Sprintf(`INSERT INTO %s(name, time) VALUES ('李雷', now())`, table))
	if err != nil {
		t.Fatal(err)
	}

	h := notifiedColumnsHandler{}
	expectedColumns := []string{"id", "time"}

	startPGNotify(notifier, table, expectedColumns, &h, t)

	_, err = db.Exec(fmt.Sprintf(`UPDATE %s SET name = '皮几万' WHERE name = '李雷'`, table))
	if err != nil {
		t.Fatal(err)
	}

	ensureNotifyReached()

	if h.pgOLD != nil || h.pgNEW != nil {
		t.Fatalf("expected: pgOLD = nil, pgNEW = nil, got pgOLD = %q, pgNEW = %q", h.pgOLD, h.pgNEW)
	}

	_, err = db.Exec(fmt.Sprintf(`UPDATE %s SET time = now() WHERE name = '皮几万'`, table))
	if err != nil {
		t.Fatal(err)
	}

	ensureNotifyReached()

	if h.pgNEW == nil || h.pgOLD == nil {
		t.Fatalf("expected: pgOLD != nil and pgNEW != nil, got pgOLD = %q, pgNEW = %q", h.pgOLD, h.pgNEW)
	}
	if bytes.Equal(h.pgOLD, h.pgNEW) {
		t.Fatalf("expected: pgOLD != pgNEW, got pgOLD = %q, pgNEW = %q", h.pgOLD, h.pgNEW)
	}

	testColumnsAsExpected(expectedColumns, h.pgNEW, t)
	testColumnsAsExpected(expectedColumns, h.pgOLD, t)
}

func testDelete(notifier *Notifier, db *sql.DB, tablePrefix string, t *testing.T) {
	table := tablePrefix + "_delete"
	createTable(db, table, t)

	h := notifiedColumnsHandler{}
	expectedColumns := []string{"name", "time"}

	_, err := db.Exec(fmt.Sprintf(`INSERT INTO %s(name, time) VALUES ('皮几万', now())`, table))
	if err != nil {
		t.Fatal(err)
	}

	startPGNotify(notifier, table, expectedColumns, &h, t)

	_, err = db.Exec(fmt.Sprintf(`DELETE FROM %s WHERE name='皮几万'`, table))
	if err != nil {
		t.Fatal(err)
	}
	ensureNotifyReached()
	testColumnsAsExpected(expectedColumns, h.pgOLD, t)
}

func ensureNotifyReached() {
	time.Sleep(100 * time.Millisecond)
}

func startPGNotify(notifier *Notifier, table string, expectedColumns []string, h Handler, t *testing.T) {
	if err := notifier.Notify(table, expectedColumns, h); err != nil {
		t.Fatal(errs.WithStack(err))
	}
}

func createTable(db *sql.DB, table string, t *testing.T) {
	_, err := db.Exec(fmt.Sprintf(`
	DROP TABLE IF EXISTS %s;
	CREATE TABLE IF NOT EXISTS %s (
		id bigserial, 
		name varchar(100),
		time timestamptz
	)`, table, table))
	if err != nil {
		t.Fatal(err)
	}
}

func testColumnsAsExpected(expectedColumns []string, notifiedMsg []byte, t *testing.T) {
	var data map[string]interface{}
	if err := json.Unmarshal(notifiedMsg, &data); err != nil {
		t.Fatal(err)
	}

	if !isColumnsAsExpected(expectedColumns, data) {
		t.Fatalf("expected columns: %q, got columns : %q", expectedColumns, keys(data))
	}
}

func isColumnsAsExpected(expectedColumns []string, notifiedMessage map[string]interface{}) bool {
	keys := make([]string, 0, len(notifiedMessage))
	if cap(keys) != len(expectedColumns) {
		return false
	}
	for _, k := range expectedColumns {
		if _, ok := notifiedMessage[k]; !ok {
			return false
		}
	}
	return true
}

func keys(d map[string]interface{}) []string {
	ks := make([]string, 0, len(d))
	for k := range d {
		ks = append(ks, k)
	}
	return ks
}

type notifiedColumnsHandler struct {
	pgNEW []byte
	pgOLD []byte
}

func (n *notifiedColumnsHandler) ConnLoss(table string) {
	// do nothing
}

func (n *notifiedColumnsHandler) Create(table string, newBuf []byte) {
	n.pgNEW = newBuf
}

func (n *notifiedColumnsHandler) Update(table string, oldBuf, newBuf []byte) {
	n.pgNEW = newBuf
	n.pgOLD = oldBuf
}

func (n *notifiedColumnsHandler) Delete(table string, oldBuf []byte) {
	n.pgOLD = oldBuf
}

type triggerActionsHandler struct {
	connLoss int
	create   int
	update   int
	delete   int
	t        *testing.T
}

func (tr *triggerActionsHandler) ConnLoss(table string) {
	tr.connLoss++
}

func (tr *triggerActionsHandler) Create(table string, newBuf []byte) {
	tr.create++
	if len(newBuf) == 0 {
		tr.t.Fatal("create does not receive record!")
	}
	tr.t.Logf("%s create: %s\n", table, newBuf)
}

func (tr *triggerActionsHandler) Update(table string, oldBuf, newBuf []byte) {
	tr.update++
	if len(oldBuf) == 0 {
		tr.t.Fatal("update does not receive old record!")
	}
	if len(newBuf) == 0 {
		tr.t.Fatal("update does not receive new record!")
	}
	tr.t.Logf("%s update: %s from: %s\n", table, newBuf, oldBuf)
}

func (tr *triggerActionsHandler) Delete(table string, oldBuf []byte) {
	tr.delete++
	if len(oldBuf) == 0 {
		tr.t.Fatal("delete does not receive record!")
	}
	tr.t.Logf("%s delete: %s\n", table, oldBuf)
}
