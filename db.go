package main

import (
	"fmt"
	"log"
	"path/filepath"

	"database/sql"
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
	if err != nil {
		log.Fatal("Error Opening SQLite Conn: ", err)
	}

	db.conn = conn
	db.name = name
	db.path = filepath.Join("db", name)

	return db
}
