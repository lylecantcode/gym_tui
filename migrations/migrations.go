package migrations

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sort"
)

// used to make sure the database is up-to date, non-reversible for now.
type Migrator struct {
	Version int
	db      *sql.DB
}

// get DB version from db migration table
func VersionCheck(db *sql.DB) {
	mig := Migrator{db: db}
	// Check current version, if no table found, value is set to 0.
	row := db.QueryRow("SELECT version FROM schema_migrations", 1)
	switch err := row.Scan(&mig.Version); err {
	case sql.ErrNoRows:
		statement, err :=
			db.Prepare("INSERT INTO schema_migrations (version) VALUES (0)")
		if err != nil {
			log.Fatalf("error with update schema_migrations statement: %s", err.Error())
		}
		_, err = statement.Exec()
		if err != nil {
			log.Fatalf("error updating schema_migrations table: %s", err.Error())
		}
	case nil:
		// log.Println("current migration", mig.Version)
	default:
		log.Println("get version error: ", err.Error())
	}
	mig.migrateUp()
}

func (m Migrator) migrateUp() {
	path := "migrations/db_files" //"README.md"
	file, err := os.Open(path)
	if err != nil {
		log.Println("opening file error: " + err.Error())
		return
	}
	defer file.Close()
	filesArray, err := file.Readdirnames(0)
	if err != nil {
		log.Println("get file names error: " + err.Error())
		return
	}
	migCount := len(filesArray) - 1
	if migCount == 0 {
		log.Println("no files found")
		return
	}
	sort.Strings(filesArray)
	filesArray = filesArray[:migCount]

	for i := m.Version; i < migCount; i++ {
		migFile := filesArray[i]
		upMig, err := os.ReadFile(fmt.Sprintf("%s/%s", path, migFile))
		if err != nil {
			log.Println("can't get migration file: ", err.Error())
			return
		}
		log.Println("migrating", migFile)
		statement, err :=
			m.db.Prepare(string(upMig))
		if err != nil {
			log.Fatalf("error with migration file %d: %s", m.Version+1, err.Error())
		}
		_, err = statement.Exec()
		if err != nil {
			log.Fatalf("error executing migration file %d: %s", m.Version+1, err.Error())
		}
		m.Version++
	}
	statement, err :=
		m.db.Prepare("UPDATE schema_migrations SET version = ?")
	if err != nil {
		log.Fatalf("error with update schema_migrations statement: %s", err.Error())
	}
	_, err = statement.Exec(m.Version)
	if err != nil {
		log.Fatalf("error updating schema_migrations table: %s", err.Error())
	}
	// apply migration and then version++
}
