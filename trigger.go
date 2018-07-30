package pgnotify

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/lovego/errs"
)

func createPGFunction(db *sql.DB) error {
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

func createTriggerIfNotExists(db *sql.DB, table string, expectedColumns []string) error {
	trigger := fmt.Sprintf("%s_pgnotify", table)
	_, err := db.Exec(fmt.Sprintf("DROP TRIGGER IF EXISTS %s ON %s", trigger, table))
	if err != nil {
		return errs.Trace(err)
	}
	columns := constructPGFuncParams(expectedColumns)
	_, err = db.Exec(fmt.Sprintf(
		`CREATE TRIGGER %s
		 AFTER INSERT OR UPDATE OR DELETE
         ON %s FOR EACH ROW
         EXECUTE PROCEDURE pgnotify(%s)`,
		trigger, table, columns))
	if err != nil {
		return errs.Trace(err)
	}
	return nil
}

func constructPGFuncParams(expectedColumns []string) string {
	if expectedColumns == nil || len(expectedColumns) == 0 {
		return `'*'`
	}
	columns := make([]string, 0, len(expectedColumns))
	for _, column := range expectedColumns {
		columns = append(columns, fmt.Sprintf(`'%s'`, column))
	}
	return strings.Join(columns, ",")
}
