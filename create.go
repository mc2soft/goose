package goose

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

// Create writes a new blank migration file.
func Create(db *sql.DB, dir, name, migrationType string) error {
	migrations, err := CollectMigrations(dir, minVersion, maxVersion)
	if err != nil {
		return err
	}

	// Initial version.
	version := "0001"

	if last, err := migrations.Last(); err == nil {
		version = fmt.Sprintf("%04v", last.Version+1)
	}

	filename := fmt.Sprintf("%v_%v.%v", version, name, migrationType)

	fpath := filepath.Join(dir, filename)
	tmpl := sqlMigrationTemplate
	if migrationType == "go" {
		tmpl = goSQLMigrationTemplate
	}

	path, err := writeTemplateToFile(fpath, tmpl, version)
	if err != nil {
		return err
	}

	fmt.Printf("Created new file: %s\n", path)
	return nil
}

func writeTemplateToFile(path string, t *template.Template, version string) (string, error) {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to create file: %v already exists", path)
	}

	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	err = t.Execute(f, version)
	if err != nil {
		return "", err
	}

	return f.Name(), nil
}

var sqlMigrationTemplate = template.Must(template.New("goose.sql-migration").Parse(`-- +goose Up
-- SQL in this section is executed when the migration is applied.

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
`))

var goSQLMigrationTemplate = template.Must(template.New("goose.go-migration").Parse(`package migrations

import (
	"database/sql"
	"github.com/mc2soft/goose"
)

func init() {
	goose.AddMigration(Up{{.}}, Down{{.}})
}

func Up{{.}}(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	_, err := tx.Exec("-- SQL goes here")
	if err != nil {
		return err
	}
	return nil
}

func Down{{.}}(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	_, err := tx.Exec("-- SQL goes here")
	if err != nil {
		return err
	}
	return nil
}
`))
