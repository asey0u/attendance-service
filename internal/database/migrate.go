package database

import (
	"database/sql"
	"os"
	"path/filepath"
	"sort"
)

func Migrate(db *sql.DB, dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.up.sql"))
	if err != nil {
		return err
	}
	sort.Strings(files)

	for _, path := range files {
		sql, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if _, err = db.Exec(string(sql)); err != nil {
			return err
		}
	}

	return nil
}
