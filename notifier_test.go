package pgnotify

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/fatih/color"
	"github.com/go-pg/pg"
	"github.com/lovego/logger"
)

func TestNotifier(t *testing.T) {
	var db = getTestDb(t)
	var notifier, err = New(db, logger.New("", os.Stderr, nil))
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond) // ensure listener has started

	testTable(notifier, db, `pgnotify_t1`, t)
	testTable(notifier, db, `pgnotify_t2`, t)

	time.Sleep(50 * time.Millisecond) // ensure event has reached
}

func testTable(notifier *Notifier, db *pg.DB, table string, t *testing.T) {
	var tbl = pg.Q(table)
	if _, err := db.Exec(`
	create table if not exists ? (
		id bigserial, name varchar(100), time timestamptz
	)`, tbl); err != nil {
		t.Fatal(err)
	}

	if err := notifier.Notify(table, testListener{}); err != nil {
		t.Fatal(err)
	}

	if _, err := db.Exec(`insert into ? (name, time) values ('李雷', now())`, tbl); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`update ? set name = '韩梅梅'`, tbl); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`delete from ?`, tbl); err != nil {
		t.Fatal(err)
	}
}

type testListener struct {
}

func (t testListener) Create(table string, buf []byte) {
	fmt.Printf("%s create: %s\n", table, buf)
}
func (t testListener) Update(table string, buf []byte) {
	fmt.Printf("%s update: %s\n", table, buf)
}
func (t testListener) Delete(table string, buf []byte) {
	fmt.Printf("%s delete: %s\n", table, buf)
}

func getTestDb(t *testing.T) *pg.DB {
	options, err := pg.ParseURL("postgres://develop:@localhost/test?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	db := pg.Connect(options)
	db.OnQueryProcessed(func(event *pg.QueryProcessedEvent) {
		query, err := event.FormattedQuery()
		if err != nil {
			log.Println(err)
		}
		log.Printf("Postgres: %s %s", time.Since(event.StartTime), color.GreenString(query))
	})
	return db
}
