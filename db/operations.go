package db

import (
	"context"
	"log"
	"time"

	"github.com/Kevin27954/liveness-sim-test/assert"
	p "github.com/Kevin27954/liveness-sim-test/pkg"
	_ "github.com/mattn/go-sqlite3"
)

func (db *DB) AddOperation(operation string, term int, message string) error {
	_, err := db.conn.Exec("INSERT INTO operations (operation, term, data) VALUES ($1, $2, $3)", operation, term, message)

	assert.NoError(err, "Error inserting operations")
	return err
}

// ID is technically the index
func (db *DB) GetLogByID(id int) (p.Operation, error) {
	row, err := db.conn.Query("SELECT * FROM operations WHERE id=$1", id)
	assert.NoError(err, "Error querying operations table by id")
	defer row.Close()

	if row.Next() {
		ops := p.Operation{}
		var timeStr string
		err := row.Scan(&ops.Id, &ops.Operation, &ops.Data, &ops.Term, &timeStr)
		if err != nil {
			log.Println("Unable to scan into p.Operation ", err)
			return p.Operation{}, err
		}

		layout := "2006-01-02 15:04:05.000"
		ops.Time, err = time.Parse(layout, timeStr)
		if err != nil {
			log.Printf("Error scanning logs: %s\n", err)
			return p.Operation{}, err
		}

		return ops, nil
	} else {
		if row.Err() != nil {
			return p.Operation{}, err
		}

		return p.Operation{}, nil
	}
}

func (db *DB) GetLogsById(startIdx int) ([]p.Operation, error) {
	rows, err := db.conn.Query("SELECT * FROM operations WHERE id>$1", startIdx)
	if err != nil {
		log.Printf("Error querying message: %s\n", err)
		return nil, err
	}
	defer rows.Close()

	// Push the logs as string into result
	opsArr := []p.Operation{}
	for rows.Next() {
		ops := p.Operation{}
		var timeStr string
		err := rows.Scan(&ops.Id, &ops.Operation, &ops.Data, &ops.Term, &timeStr)
		if err != nil {
			log.Println("Unable to scan into p.Operation, ", err)
			return []p.Operation{}, err
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
		return []p.Operation{}, err
	}

	return opsArr, nil
}

// Might be pointless. Use getlogsbyid
func (db *DB) GetLogs() ([]p.Operation, error) {
	rows, err := db.conn.Query("SELECT * FROM operations")
	defer rows.Close()
	if err != nil {
		log.Printf("Error querying message: %s\n", err)
		return nil, err
	}

	var operations []p.Operation

	for rows.Next() {
		var op p.Operation
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

	// for _, op := range operations {
	// 	log.Println("OP:", op)
	// }

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

func (db *DB) HasLog(ops p.Operation) (bool, error) {
	row, err := db.conn.Query("SELECT * FROM operations WHERE id=$1 AND operation=$2 AND term=$3 AND data=$4", ops.Id, ops.Operation, ops.Term, ops.Data)
	if err != nil {
		return false, err
	}
	defer row.Close()

	return row.Next(), nil
}

// I need to look into this more
func (db *DB) CommitLogs(opsArr []p.Operation) (bool, error) {
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
