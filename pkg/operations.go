package pkg

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
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
