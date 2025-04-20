package db

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Kevin27954/liveness-sim-test/assert"
	_ "github.com/mattn/go-sqlite3"
)

type Operation struct {
	Id        int
	Operation string
	Time      time.Time
	Term      int
	Data      string
}

func (o Operation) String() string {
	seperator := "*"                                    // Some random seperator
	layout := "2006-01-02 15:04:05.999999999 -0700 MST" // format used for time

	return fmt.Sprintf("%d%s%s%s%s%s%d%s%s", o.Id, seperator, o.Operation, seperator, o.Time.Format(layout), seperator, o.Term, seperator, o.Data)
}

func (db *DB) AddOperation(operation string, term int, message string) {
	_, err := db.conn.Exec(fmt.Sprintf("INSERT INTO operations (operation, term, data) VALUES ('%s', '%d', '%s')", operation, term, message))
	assert.NoError(err, "Error inserting operations")
	log.Println("succesffully added operations")
}

// ID is technically the index
func (db *DB) GetLogByID(id int) (Operation, error) {

	return Operation{}, nil
}

func (db *DB) GetMissingLogs(startIdx int) ([]Operation, error) {
	// rows, err := db.conn.Query("SELECT * FROM operations WHERE id>startIdx")
	// defer rows.Close()
	// if err != nil {
	// 	log.Printf("Error querying message: %s\n", err)
	// 	return nil, fmt.Errorf("Error scanning message")
	// }
	//
	// Push the logs as string into result

	return []Operation{}, nil
}

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
		err := rows.Scan(&op.Id, &op.Operation, &op.Data, &timeStr)
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
	// WRITE CODE TO:
	// PARSE logStr
	// CHECK DB (MAYBE DO BLOOM FILTER TOO IF I HAVE TIME OR INTEREST)
	// IF RESULT IS NOT === 0 RETURN TRUE
	// ELSE RETURN FALSE
	// OF CORUSE RETURN ANY ERRORS

	return true, nil
}

// Helper Function to Parse Logs from String to Log
func parseLogs(logStr string) (Operation, error) {
	data := strings.Split(logStr, "*")
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
