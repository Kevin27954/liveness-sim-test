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

type Message struct {
	Id  int
	Msg string
}

type Operation struct {
	Operation string
	// Msgid     int
	Msg string
}

func Init(name string) DB {
	db := DB{}

	conn, err := sql.Open("sqlite3", filepath.Join("sqlite_file", fmt.Sprintf("%s.db", name)))
	assert.NoError(err, "Error Opening SQLite Conn")

	db.conn = conn
	db.name = name
	db.path = filepath.Join("db", name)

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
	// data can be Something else other than Text
	// For now we can ignore the ID part. I just wnt it working for now.
	// _, err := db.conn.Exec("CREATE TABLE IF NOT EXISTS operations (id INTEGER PRIMARY KEY, operation TEXT, msgid INTEGER, data TEXT, FOREIGN KEY(msgid) REFERENCES message(id))")

	// Time is actually SUPER IMPORTANT HERE

	_, err := db.conn.Exec("CREATE TABLE IF NOT EXISTS operations (id INTEGER PRIMARY KEY, operation TEXT, data TEXT)")
	assert.NoError(err, "Error Creating Operations Table")
}

func (db *DB) AddMessage(message string) {
	_, err := db.conn.Exec(fmt.Sprintf("INSERT INTO message (message) VALUES ('%s')", message))
	assert.NoError(err, "Error inserting message")
}

func (db *DB) GetMessages() ([]Message, error) {
	rows, err := db.conn.Query("SELECT * FROM message")
	defer rows.Close()
	if err != nil {
		log.Printf("Error querying message: %s\n", err)
		return nil, fmt.Errorf("Error scanning message")
	}

	var messages []Message

	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.Id, &msg.Msg)

		if err != nil {
			log.Printf("Error scanning message: %s\n", err)
			return nil, fmt.Errorf("Error scanning messages")
		}

		messages = append(messages, msg)
	}

	assert.NoError(rows.Err(), "Error iterating message rows")
	return messages, nil
}

func (db *DB) AddOperation(operation string, message string) {
	_, err := db.conn.Exec(fmt.Sprintf("INSERT INTO operations (operation, data) VALUES ('%s', '%s')", operation, message))
	assert.NoError(err, "Error inserting operations")
}

func (db *DB) GetMissingLogs(startIdx int) ([]string, error) {
	rows, err := db.conn.Query("SELECT * FROM operations WHERE id>startIdx")
	defer rows.Close()
	if err != nil {
		log.Printf("Error querying message: %s\n", err)
		return nil, fmt.Errorf("Error scanning message")
	}

	return []string{}, nil
}
