package db

import (
	"fmt"
	"log"
	"path/filepath"

	"database/sql"

	"github.com/Kevin27954/liveness-sim-test/assert"
	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	name string
	path string
	conn *sql.DB
}

func Init(name string) DB {
	db := DB{}

	conn, err := sql.Open("sqlite3", filepath.Join("sqlite_file", fmt.Sprintf("%s.db?_journal_mode=WAL&synchronous=normal", name)))
	assert.NoError(err, "Error Opening SQLite Conn")

	db.conn = conn
	db.name = name
	db.path = filepath.Join("db", name)

	_, err = conn.Exec("PRAGMA optimize")
	if err != nil {
		log.Fatal("Error:", err)
	}

	assert.NotNil(db.conn, "db conn is nil")
	db.createMessageTable()
	db.createOperationTable()

	return db
}

func (db *DB) createMessageTable() {
	_, err := db.conn.Exec("CREATE TABLE IF NOT EXISTS message (id INTEGER PRIMARY KEY, message TEXT)")
	assert.NoError(err, "Error Creating Message Table")
}

func (db *DB) createOperationTable() {
	_, err := db.conn.Exec("CREATE TABLE IF NOT EXISTS operations (id INTEGER PRIMARY KEY, operation TEXT, data TEXT, term INTEGER, time INTEGER DEFAULT (DATETIME('now', 'subsec')))")
	assert.NoError(err, "Error Creating Operations Table")
}

// This needs to be after close. But we only close with +c, so how do we call this function then? Perhaps a http call to the thing
// which closes the thing here?
func (db *DB) Close() {
	db.conn.Close()
}
