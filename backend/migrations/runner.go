package migrations

import (
	"database/sql"
	"embed"
	"io/fs"
)

// main function for running the migration
func Run(db *sql.DB) error {
	err := createTable(db)
	if err != nil {
		return err
	}
	migrations, err := getAppliedMigrations(db)
	if err != nil {
		return err
	}
	allFiles, err := listMigrationFiles(MigrationsFS)
	if err != nil {
		return err
	}
	newVersions := findNewVersions(migrations, allFiles)
	for _, name := range newVersions {
		query, err := readMigrationSQL(MigrationsFS, name)
		if err != nil {
			return err
		}
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()
		if _, err := tx.Exec(query); err != nil {
			return err
		}
		newQuery := `INSERT INTO schema_migrations (version) VALUES (?)`
		if _, err := tx.Exec(newQuery, name); err != nil {
			return err
		}
		if err = tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

func createTable(db *sql.DB) error {
	query := `CREATE TABLE IF NOT EXISTS schema_migrations(
    version TEXT PRIMARY KEY,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
);`
	_, err := db.Exec(query)
	return err
}

// get the map of already existing versions
func getAppliedMigrations(db *sql.DB) (map[string]bool, error) {
	migrations := make(map[string]bool)
	query := `SELECT version FROM schema_migrations`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		version := ""
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		migrations[version] = true
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return migrations, nil
}

// open all migrations files
func listMigrationFiles(migsFS embed.FS) ([]string, error) {
	listMigrations := []string{}
	listOfFiles, err := fs.ReadDir(migsFS, ".")
	if err != nil {
		return nil, err
	}
	for _, entry := range listOfFiles {
		name := entry.Name()
		listMigrations = append(listMigrations, name)
	}
	return listMigrations, nil
}

// function that check versions that weren't applied yet
func findNewVersions(appliedVersions map[string]bool, allFiles []string) []string {
	newVersion := []string{}
	for _, version := range allFiles {
		if !appliedVersions[version] {
			newVersion = append(newVersion, version)
		}
	}
	return newVersion
}

// function that take the name of file and a file and return the text inside file
func readMigrationSQL(file embed.FS, name string) (string, error) {
	text, err := fs.ReadFile(file, name)
	if err != nil {
		return "", err
	}
	return string(text), nil
}
