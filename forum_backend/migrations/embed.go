package migrations

import (
	"embed"
)

//go:embed *.sql
var MigrationsFS embed.FS

func ListMigrationFiles(migsFS embed.FS) []string{
	listMigrations:=[]string{}
	listOfFiles, err:= fs.
	return listMigrations
}