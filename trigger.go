package pgnotify

import (
	"github.com/go-pg/pg"
	"github.com/lovego/errs"
)

func CreateFunction(db *pg.DB) error {
	if _, err := db.Exec(`create or replace function pgnotify() returns trigger as $$
	begin
		perform pg_notify('pgnotify_' || tg_table_name, json_build_object(
			'action', tg_op,
			'data', row_to_json(case when tg_op = 'DELETE' then old else new end)
		)::text);
		return null;
	end;
$$ language plpgsql`); err != nil {
		return errs.Trace(err)
	}
	return nil
}

func CreateTriggerIfNotExists(db *pg.DB, table string) error {
	var count int
	if _, err := db.Query(pg.Scan(&count),
		`select count(*) as count from pg_trigger
		where tgrelid = ?::regclass and tgname ='?_pgnotify' and not tgisinternal`,
		table, pg.Q(table),
	); err != nil {
		return errs.Trace(err)
	}
	if count <= 0 {
		if _, err := db.Exec(
			`create trigger ?_pgnotify after insert or update or delete on ?
			for each row execute procedure pgnotify()`,
			pg.Q(table), pg.Q(table),
		); err != nil {
			return errs.Trace(err)
		}
	}
	return nil
}
