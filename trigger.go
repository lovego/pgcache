package pgnotify

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/lovego/errs"
)

func createPGFunction(db *sql.DB) error {
	// tg_argv[0] 是需要通知的字段列表
	// tg_argv[1] 是需要检查是否有变动的字段列表，仅在更新时使用
	_, err := db.Exec(`
    create or replace function pgnotify() returns trigger as $$
    declare
      old_record record;
      new_record record;
      data jsonb;
    begin
      if tg_op = 'UPDATE' and tg_argv[1] != '' then
        execute 'select ' || tg_argv[1] into old_record using old;
        execute 'select ' || tg_argv[1] into new_record using new;
        if old_record = new_record then
          return null;
        end if;
      end if;

      data := json_build_object('action', tg_op);
      case tg_op
      when 'INSERT' then
        execute 'select ' || tg_argv[0] into new_record using new;
        data := jsonb_set(data, array['new'], to_jsonb(new_record));
      when 'UPDATE' then
        execute 'select ' || tg_argv[0] into old_record using old;
        execute 'select ' || tg_argv[0] into new_record using new;
        data := jsonb_set(data, array['old'], to_jsonb(old_record));
        data := jsonb_set(data, array['new'], to_jsonb(new_record));
      when 'DELETE' then
        execute 'select ' || tg_argv[0] into old_record using old;
        data := jsonb_set(data, array['old'], to_jsonb(old_record));
      end case;

      perform pg_notify('pgnotify_' || tg_table_name, data::text);
      return null;
    end;
    $$ language plpgsql;`)
	if err != nil {
		return errs.Trace(err)
	}
	return nil
}

func createTrigger(db *sql.DB, table string, columnsToNotify, columnsToCheck string) error {
	trigger := fmt.Sprintf("%s_pgnotify", table)
	_, err := db.Exec(fmt.Sprintf("DROP TRIGGER IF EXISTS %s ON %s", trigger, table))
	if err != nil {
		return errs.Trace(err)
	}
	if _, err = db.Exec(fmt.Sprintf(
		`CREATE TRIGGER %s
		 AFTER INSERT OR UPDATE OR DELETE
         ON %s FOR EACH ROW
         EXECUTE PROCEDURE pgnotify(%s, %s)`,
		trigger, table, quote(columnsToNotify), quote(columnsToCheck)),
	); err != nil {
		return errs.Trace(err)
	}
	return nil
}

func quote(q string) string {
	return "'" + strings.Replace(q, "'", "''", -1) + "'"
}
