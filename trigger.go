package pgnotify

import (
	"database/sql"
	"fmt"

	"github.com/lovego/errs"
)

func CreateFunction(db *sql.DB) error {
	var count int
	if err := db.QueryRow(
		`select count(*) as count from pg_proc where proname ='pgnotify'`,
	).Scan(&count); err != nil {
		return errs.Trace(err)
	}
	if count > 0 {
		return nil
	}
	if _, err := db.Exec(`create or replace function pgnotify() returns trigger as $$
	begin
		perform pg_notify('pgnotify_' || tg_table_name, json_build_object(
			'action', tg_op,
			'data', row_to_json(case when tg_op = 'DELETE' then old else new end),
      'old', row_to_json(case when tg_op = 'UPDATE' then old end)
		)::text);
		return null;
	end;
$$ language plpgsql`); err != nil {
		return errs.Trace(err)
	}
	return nil
}

func CreateTriggerIfNotExists(db *sql.DB, table string) error {
	var count int
	if err := db.QueryRow(
		`select count(*) as count from pg_trigger
		where tgrelid = $1::regclass and tgname = $2 and not tgisinternal`,
		table, table+"_pgnotify",
	).Scan(&count); err != nil {
		return errs.Trace(err)
	}
	if count > 0 {
		return nil
	}
	if _, err := db.Exec(fmt.Sprintf(
		`create trigger %s_pgnotify after insert or update or delete on %s
			for each row execute procedure pgnotify()`, table, table,
	)); err != nil {
		return errs.Trace(err)
	}
	return nil
}
