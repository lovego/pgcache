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
	if err := notifier.Notify(table, &h); err != nil {
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
	if h.create != 1 || h.update != 1 && h.delete != 1 {
		t.Errorf("expected: %+v", h)
	}
}

type testHandler struct {
	create, update, delete int
	t                      *testing.T
}

func (h *testHandler) Create(table string, buf []byte) {
	h.create++
	h.t.Logf("%s create: %s\n", table, buf)
}
func (h *testHandler) Update(table string, buf []byte) {
	h.update++
	h.t.Logf("%s update: %s\n", table, buf)
}
func (h *testHandler) Delete(table string, buf []byte) {
	h.delete++
	h.t.Logf("%s delete: %s\n", table, buf)
}
