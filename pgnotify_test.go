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

	if _, err := db.Exec(`
	drop table if exists pgnotify_tests;
	create table pgnotify_tests (
		id bigserial,
		name varchar(100),
		time timestamptz
	)`); err != nil {
		t.Fatal(err)
	}

	var ln = testListener{}
	var notifier = Notifier{
		DB: db,
		Listeners: map[string]Listener{
			`pgnotify_tests`: &ln,
		},
		Logger: logger.New("", os.Stderr, nil),
	}
	if err := notifier.Start(); err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Second)

	if _, err := db.Exec(`
	insert into pgnotify_tests (name, time) values ('李雷', now())
	`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`
		update pgnotify_tests set name = '韩梅梅'
		`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`
		delete from pgnotify_tests where name = '韩梅梅'
		`); err != nil {
		t.Fatal(err)
	}
	time.Sleep(5 * time.Second)
}

type testListener struct {
}

func (t *testListener) Create(buf []byte) {
	fmt.Printf("create: %s\n", buf)
}
func (t *testListener) Update(buf []byte) {
	fmt.Printf("update: %s\n", buf)
}
func (t *testListener) Delete(buf []byte) {
	fmt.Printf("delete: %s\n", buf)
}
