package pgnotify

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/lovego/errs"
)

func CreateFunction(db *sql.DB) error {
	_, err := db.Exec(`
        CREATE OR REPLACE FUNCTION pgnotify() RETURNS TRIGGER AS $$
        DECLARE
            field TEXT;
            query TEXT;
            new_row RECORD;
            old_row RECORD;
            new_json JSONB := '{}';
            old_json JSONB := '{}';
        BEGIN
            query := 'SELECT ';
            FOREACH field IN ARRAY TG_ARGV LOOP
                query := query || format('($1).%s',  field) || ',';
            END LOOP;
            query := rtrim(query, ',');

            IF TG_OP = 'UPDATE' OR TG_OP = 'DELETE' THEN
                EXECUTE query
                INTO old_row
                USING OLD;
                old_json := row_to_json(old_row);
            END IF;

            IF TG_OP = 'UPDATE' OR TG_OP = 'INSERT' THEN
                EXECUTE query
                INTO new_row
                USING NEW;
                new_json := row_to_json(new_row);  
            END IF;

            IF (TG_OP = 'UPDATE' AND new_json <> old_json) OR TG_OP = 'INSERT' OR TG_OP = 'DELETE' THEN
                PERFORM pg_notify('pgnotify_' || TG_TABLE_NAME, json_build_object(
                    'action', TG_OP,
                    'old', CASE WHEN TG_OP != 'INSERT' THEN old_json END,
                    'new', CASE WHEN TG_OP != 'DELETE' THEN new_json END
                )::TEXT);
            END IF;

            RETURN NULL;
        END;
        $$ LANGUAGE PLPGSQL;`)
	if err != nil {
		return errs.Trace(err)
	}
	return nil
}

func CreateTriggerIfNotExists(db *sql.DB, table string, wantedColumns []string) error {
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

	columns := constructPGFuncParams(wantedColumns)
	if _, err := db.Exec(fmt.Sprintf(
		`create trigger %s_pgnotify after insert or update or delete on %s
			for each row execute procedure pgnotify(%s)`, table, table, columns,
	)); err != nil {
		return errs.Trace(err)
	}
	return nil
}

func constructPGFuncParams(wantedColumns []string) string {
	if wantedColumns == nil || len(wantedColumns) == 0 {
		return `'*'`
	}
	columns := make([]string, 0, len(wantedColumns))
	for _, column := range wantedColumns {
		columns = append(columns, fmt.Sprintf(`'%s'`, column))
	}
	return strings.Join(columns, ",")
}
