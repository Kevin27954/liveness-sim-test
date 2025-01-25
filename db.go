package main

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

	conn, err := sql.Open("sqlite3", filepath.Join("db", fmt.Sprintf("%s.db", name)))
	assert.NoError(err, "Error Opening SQLite Conn")

	db.conn = conn
	db.name = name
	db.path = filepath.Join("db", name)

	assert.NotNil(db.conn, "db conn is nil")
	db.createMessageTable()

	return db
}

func (db *DB) createMessageTable() {
	_, err := db.conn.Exec("CREATE TABLE IF NOT EXISTS message (id INTEGER PRIMARY KEY, message TEXT)")
	assert.NoError(err, "Error Creating Table")
}

func (db *DB) AddMessage(message string) {
	_, err := db.conn.Exec(fmt.Sprintf("INSERT INTO message (message) VALUES ('%s')", message))
	assert.NoError(err, "Error inserting message")
}

func (db *DB) GetMessages() ([]string, error) {
	rows, err := db.conn.Query("SELECT * FROM message")
	defer rows.Close()
	if err != nil {
		log.Printf("Error querying message: %s\n", err)
		return nil, fmt.Errorf("Error scanning message")
	}

	var messages []string
	var message string
	var id int

	for rows.Next() {
		err := rows.Scan(&id, &message)

		if err != nil {
			log.Printf("Error scanning message: %s\n", err)
			return nil, fmt.Errorf("Error scanning messages")
		}

		messages = append(messages, message)
	}

	assert.NoError(rows.Err(), "Error iterating message rows")
	return messages, nil
}
