package migrations

import "database/sql"

func CreateTable(db *sql.DB) error {
	query := `CREATE TABLE IF NOT EXISTS schema_migrations(
    version TEXT PRIMARY KEY,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
);`
	_, err := db.Exec(query)
	return err
}
func GetAppliedMigrations(db *sql.DB) (map[string]bool, error) {
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
