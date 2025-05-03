package db

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Kevin27954/liveness-sim-test/assert"
	_ "github.com/mattn/go-sqlite3"
)

const SEPERATOR = "◊"

type Operation struct {
	Id        int
	Operation string
	Time      time.Time
	Term      int
	Data      string
}

func (o Operation) String() string {
	layout := "2006-01-02 15:04:05.999999999 -0700 MST" // format used for time

	if o.Data == "" && o.Operation == "" {
		return "Empty"
	}

	return fmt.Sprintf("%d%s%s%s%s%s%d%s%s", o.Id, SEPERATOR, o.Operation, SEPERATOR, o.Time.Format(layout), SEPERATOR, o.Term, SEPERATOR, o.Data)
}

func (db *DB) AddOperation(operation string, term int, message string) error {
	_, err := db.conn.Exec("INSERT INTO operations (operation, term, data) VALUES ($1, $2, $3)", operation, term, message)

	assert.NoError(err, "Error inserting operations")
	return err
}

// ID is technically the index
func (db *DB) GetLogByID(id int) (Operation, error) {
	row, err := db.conn.Query("SELECT * FROM operations WHERE id=$1", id)
	assert.NoError(err, "Error querying operations table by id")
	defer row.Close()

	if row.Next() {
		ops := Operation{}
		var timeStr string
		err := row.Scan(&ops.Id, &ops.Operation, &ops.Data, &ops.Term, &timeStr)
		if err != nil {
			log.Println("Unable to scan into Operation ", err)
			return Operation{}, err
		}

		layout := "2006-01-02 15:04:05.000"
		ops.Time, err = time.Parse(layout, timeStr)
		if err != nil {
			log.Printf("Error scanning logs: %s\n", err)
			return Operation{}, err
		}

		return ops, nil
	} else {
		if row.Err() != nil {
			return Operation{}, err
		}

		return Operation{}, nil
	}
}

func (db *DB) GetLogsById(startIdx int) ([]Operation, error) {
	rows, err := db.conn.Query("SELECT * FROM operations WHERE id>$1", startIdx)
	if err != nil {
		log.Printf("Error querying message: %s\n", err)
		return nil, err
	}
	defer rows.Close()

	// Push the logs as string into result
	opsArr := []Operation{}
	for rows.Next() {
		ops := Operation{}
		var timeStr string
		err := rows.Scan(&ops.Id, &ops.Operation, &ops.Data, &ops.Term, &timeStr)
		if err != nil {
			log.Println("Unable to scan into Operation, ", err)
			return []Operation{}, err
		}

		layout := "2006-01-02 15:04:05.000"
		ops.Time, err = time.Parse(layout, timeStr)
		if err != nil {
			log.Printf("Error scanning logs: %s\n", err)
			return nil, err
		}

		opsArr = append(opsArr, ops)
	}

	if rows.Err() != nil {
		return []Operation{}, err
	}

	return opsArr, nil
}

// Might be pointless. Use getlogsbyid
func (db *DB) GetLogs() ([]Operation, error) {
	rows, err := db.conn.Query("SELECT * FROM operations")
	defer rows.Close()
	if err != nil {
		log.Printf("Error querying message: %s\n", err)
		return nil, err
	}

	var operations []Operation

	for rows.Next() {
		var op Operation
		var timeStr string
		err := rows.Scan(&op.Id, &op.Operation, &op.Data, &op.Term, &timeStr)
		if err != nil {
			log.Printf("Error scanning logs: %s\n", err)
			return nil, err
		}

		layout := "2006-01-02 15:04:05.000"
		op.Time, err = time.Parse(layout, timeStr)
		if err != nil {
			log.Printf("Error scanning logs: %s\n", err)
			return nil, err
		}

		operations = append(operations, op)
	}

	assert.NoError(rows.Err(), "Error iterating message rows")

	for _, op := range operations {
		log.Println("OP:", op)
	}

	return operations, nil
}

func (db *DB) GetNumLogs() (int, error) {
	num, err := db.conn.Query("SELECT count(*) FROM operations")
	defer num.Close()
	if err != nil {
		log.Printf("Error querying # of operations: %s\n", err)
		return 0, err
	}

	if num.Next() {
		var numRow int
		err = num.Scan(&numRow)
		if err != nil {
			log.Printf("Error scanningo # of perations logs: %s\n", err)
			return 0, err
		}

		return numRow, nil
	} else {
		return 0, err
	}
}

func (db *DB) HasLog(ops Operation) (bool, error) {
	row, err := db.conn.Query("SELECT * FROM operations WHERE id=$1 AND operation=$2 AND term=$3 AND data=$4", ops.Id, ops.Operation, ops.Term, ops.Data)
	if err != nil {
		return false, err
	}
	defer row.Close()

	return row.Next(), nil
}

// I need to look into this more
func (db *DB) CommitLogs(opsArr []Operation) (bool, error) {
	hasErr := false

	tx, err := db.conn.BeginTx(context.Background(), nil)
	if err != nil {
		return false, err
	}

	_, err = tx.Exec("DELETE FROM operations WHERE id>=$1", opsArr[0].Id)
	if err != nil {
		log.Println("Error performing TX - Delete Op: ", err)
	}

	stmt, err := tx.Prepare("INSERT INTO operations (operation, term, data) VALUES (?, ?, ?)")
	defer stmt.Close()

	for _, ops := range opsArr {
		_, err = stmt.Exec(ops.Operation, ops.Term, ops.Data)
		if err != nil {
			hasErr = true
			log.Println("Error exec TX STMT: ", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		hasErr = true
	}

	if hasErr {
		log.Println("Error from Commiting: ", err)
		err := tx.Rollback()
		if err != nil {
			log.Println("Unable to rollback: ", err)
		}
	}

	return true, nil
}

// Helper Function to Parse Logs from String to Log
func ParseLog(logStr string) (Operation, error) {
	if logStr == "" {
		return Operation{}, nil
	}

	data := strings.Split(logStr, SEPERATOR)
	op := Operation{}
	var err error

	format := "2006-01-02 15:04:05.999999999 -0700 MST"

	op.Id, err = strconv.Atoi(data[0])
	op.Operation = data[1]
	op.Time, err = time.Parse(format, data[2])
	op.Term, err = strconv.Atoi(data[3])
	op.Data = strings.Join(data[3:], "*")

	if err != nil {
		log.Println("Unable to convert string to Operation. Check Id, Time, and Term value: ", err, "\n", logStr)
		return Operation{}, err
	}

	return op, nil
}
