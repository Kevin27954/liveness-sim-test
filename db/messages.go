package db

import (
	"log"

	"github.com/Kevin27954/liveness-sim-test/assert"
	_ "github.com/mattn/go-sqlite3"
)

type Message struct {
	Id  int
	Msg string
}

func (db *DB) AddMessage(message string) error {
	_, err := db.conn.Exec("INSERT INTO message (message) VALUES ($1)", message)
	assert.NoError(err, "Error inserting message")
	return err
}

func (db *DB) GetMessages() ([]Message, error) {
	rows, err := db.conn.Query("SELECT * FROM message")
	defer rows.Close()
	if err != nil {
		log.Printf("Error querying message: %s\n", err)
		return nil, err
	}

	var messages []Message

	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.Id, &msg.Msg)

		if err != nil {
			log.Printf("Error scanning message: %s\n", err)
			return nil, err
		}

		messages = append(messages, msg)
	}

	assert.NoError(rows.Err(), "Error iterating message rows")
	return messages, nil
}
