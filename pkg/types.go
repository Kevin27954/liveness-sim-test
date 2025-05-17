package pkg

import (
	"log"
	"strconv"
	"strings"
)

type TickerEvent struct {
	Event    string
	From     int
	Contents string
}

type MessageEvent struct {
	Event    string
	From     int
	Contents string
}

// Return any (Up to the parse thing at that spot to figure out from the swithc
func Init(msg string) any {
	seperator := "≡"

	data := strings.Split(msg, seperator)
	from, err := strconv.Atoi(data[1]) //from
	if err != nil {
		log.Fatal("Unable to parse from", err)
	}

	if data[0] == HEARTBEAT {
		return TickerEvent{
			Event:    data[0],
			From:     from,
			Contents: strings.Join(data[2:], seperator),
		}
	}

	m := MessageEvent{
		Event:    data[0],
		From:     from,
		Contents: strings.Join(data[2:], seperator),
	}

	return m
}
