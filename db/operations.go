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

func (db *DB) AddOperation(operation string, term int, message string) {
	_, err := db.conn.Exec("INSERT INTO operations (operation, term, data) VALUES ($1, $2, $3)", operation, term, message)
	assert.NoError(err, "Error inserting operations")
	log.Println("succesffully added operations")
}

// ID is technically the index
func (db *DB) GetLogByID(id int) (Operation, error) {
	row, err := db.conn.Query("SELECT * FROM operations WHERE id=$1", id)
	assert.NoError(err, "Error querying operations table by id")

	if row.Next() {
		ops := Operation{}
		var timeStr string
		err := row.Scan(&ops.Id, &ops.Operation, &ops.Data, &ops.Term, &timeStr)
		if err != nil {
			log.Println("Unable to scan into Operation ", err)
			return Operation{}, fmt.Errorf("Unable or no data to get for Row data in GetLogById()")
		}

		layout := "2006-01-02 15:04:05.000"
		ops.Time, err = time.Parse(layout, timeStr)
		if err != nil {
			log.Printf("Error scanning logs: %s\n", err)
			return Operation{}, fmt.Errorf("Error parsing time in logs")
		}

		return ops, nil
	} else {
		if row.Err() != nil {
			return Operation{}, fmt.Errorf("Unable or no data to get for Row data in GetLogById()")
		}

		return Operation{}, nil
	}
}

func (db *DB) GetLogsById(startIdx int) ([]Operation, error) {
	rows, err := db.conn.Query("SELECT * FROM operations WHERE id>$1", startIdx)
	defer rows.Close()
	if err != nil {
		log.Printf("Error querying message: %s\n", err)
		return nil, fmt.Errorf("Error scanning message")
	}

	// Push the logs as string into result
	opsArr := []Operation{}
	for rows.Next() {
		ops := Operation{}
		var timeStr string
		err := rows.Scan(&ops.Id, &ops.Operation, &ops.Data, &ops.Term, &timeStr)
		if err != nil {
			log.Println("Unable to scan into Operation")
			return []Operation{}, fmt.Errorf("Unable to scan into operation: %s", err)
		}

		layout := "2006-01-02 15:04:05.000"
		ops.Time, err = time.Parse(layout, timeStr)
		if err != nil {
			log.Printf("Error scanning logs: %s\n", err)
			return nil, fmt.Errorf("Error parsing time in logs")
		}

		opsArr = append(opsArr, ops)
	}

	if rows.Err() != nil {
		return []Operation{}, fmt.Errorf("Unable to get missing logs: %s", rows.Err())
	}

	return opsArr, nil
}

// Might be pointless. Use getlogsbyid
func (db *DB) GetLogs() ([]Operation, error) {
	rows, err := db.conn.Query("SELECT * FROM operations")
	defer rows.Close()
	if err != nil {
		log.Printf("Error querying message: %s\n", err)
		return nil, fmt.Errorf("Error getting logs")
	}

	var operations []Operation

	for rows.Next() {
		var op Operation
		var timeStr string
		err := rows.Scan(&op.Id, &op.Operation, &op.Data, &op.Term, &timeStr)
		if err != nil {
			log.Printf("Error scanning logs: %s\n", err)
			return nil, fmt.Errorf("Error scanning logs")
		}

		layout := "2006-01-02 15:04:05.000"
		op.Time, err = time.Parse(layout, timeStr)
		if err != nil {
			log.Printf("Error scanning logs: %s\n", err)
			return nil, fmt.Errorf("Error parsing time in logs")
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
		return 0, fmt.Errorf("Error getting count of logs")
	}

	if num.Next() {
		var numRow int
		err = num.Scan(&numRow)
		if err != nil {
			log.Printf("Error scanningo # of perations logs: %s\n", err)
			return 0, fmt.Errorf("Error scanning logs")
		}

		return numRow, nil
	} else {
		return 0, fmt.Errorf("Error scanning logs")
	}
}

func (db *DB) HasLog(logStr string) (bool, error) {
	ops, err := parseLog(logStr)
	if err != nil {
		return false, err
	}

	row, err := db.conn.Query("SELECT * FROM operations WHERE id=$1 AND operation=$2 AND term=$3", ops.Id, ops.Operation, ops.Term)
	if err != nil {
		return false, err
	}

	return row.Next(), nil
}

func (db *DB) CommitLogs(missingLogs []string) (bool, error) {

	log.Println("ALL LOGS: ", missingLogs, " LENTH: ", len(missingLogs))
	log.Println("The value: ", missingLogs[0], " is equal to empty? ", missingLogs[0] == "")

	// if len(missingLogs) == 0 || len(missingLogs) == 1 && len(missingLogs[0]) == 0 {
	if len(missingLogs) == 0 {
		log.Println("Was empty")
		return true, nil
	}

	var opsArr []Operation
	hasErr := false

	for _, ops := range missingLogs {
		ops, err := parseLog(ops)
		if err != nil {
			return false, err
		}

		opsArr = append(opsArr, ops)
	}

	tx, err := db.conn.BeginTx(context.Background(), nil)
	if err != nil {
		return false, err
	}

	_, err = tx.Exec("DELETE FROM operations WHERE id>=$1", opsArr[0].Id)
	if err != nil {
		log.Fatal("FUCKED UP, ", err)
	}

	stmt, err := tx.Prepare("INSERT INTO operations (operation, term, data) VALUES (?, ?, ?)")
	defer stmt.Close()

	for _, ops := range opsArr {
		_, err = stmt.Exec(ops.Operation, ops.Term, ops.Data)
		if err != nil {
			hasErr = true
			log.Println("error happend during stmt exec in tx: ", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		hasErr = true
	}

	if hasErr {
		log.Println("Error from commiting: ", err)
		err := tx.Rollback()
		if err != nil {
			log.Println("WTF, Can't rollback: ", err)
		}
	}

	log.Println("I finished")

	return true, nil
}

// Helper Function to Parse Logs from String to Log
func parseLog(logStr string) (Operation, error) {
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
