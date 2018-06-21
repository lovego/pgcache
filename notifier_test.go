package pgnotify

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/lovego/errs"
	"github.com/lovego/logger"
)

func TestNotifier(t *testing.T) {
	var addr = "postgres://develop:@localhost/test?sslmode=disable"
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

	testTable(notifier, db, `pgnotify_t1`, t)
	testTable(notifier, db, `pgnotify_t2`, t)

}

func testTable(notifier *Notifier, db *sql.DB, table string, t *testing.T) {
	if _, err := db.Exec(`
	drop table if exists ` + table + `;
	create table if not exists ` + table + ` (
		id bigserial, name varchar(100), time timestamptz
	)`); err != nil {
		t.Fatal(err)
	}

	h := testHandler{t: t}
	if err := notifier.Notify(table, []string{"time", "id", "name"}, &h); err != nil {
		t.Fatal(errs.WithStack(err))
	}

	if _, err := db.Exec(`insert into ` + table + ` (name, time) values ('李雷', now())`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`update ` + table + ` set name = '韩梅梅'`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`delete from ` + table); err != nil {
		t.Fatal(err)
	}

	time.Sleep(100 * time.Millisecond) // ensure event has reached
	if h.connLoss != 1 || h.create != 1 || h.update != 1 && h.delete != 1 {
		t.Errorf("expected: %+v", h)
	}
}

type testHandler struct {
	connLoss, create, update, delete int
	t                                *testing.T
}

func (h *testHandler) ConnLoss(table string) {
	h.connLoss++
}

func (h *testHandler) Create(table string, newBuf []byte) {
	h.create++
	if len(newBuf) == 0 {
		h.t.Fatal("create does not receive record!")
	}
	h.t.Logf("%s create: %s\n", table, newBuf)
}
func (h *testHandler) Update(table string, oldBuf, newBuf []byte) {
	h.update++
	if len(oldBuf) == 0 {
		h.t.Fatal("update does not receive old record!")
	}
	if len(newBuf) == 0 {
		h.t.Fatal("update does not receive new record!")
	}
	h.t.Logf("%s update: %s from: %s\n", table, newBuf, oldBuf)
}
func (h *testHandler) Delete(table string, oldBuf []byte) {
	h.delete++
	if len(oldBuf) == 0 {
		h.t.Fatal("delete does not receive record!")
	}
	h.t.Logf("%s delete: %s\n", table, oldBuf)
}
