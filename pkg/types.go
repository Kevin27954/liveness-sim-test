package raft

import (
	"log"
	"strconv"
	"strings"
)

type TickerEvent struct{}

type MessageEvent struct {
	From     int
	Event    string
	Contents string
}

func Init(msg string) MessageEvent {
	seperator := "≡"

	data := strings.Split(msg, seperator)
	from, err := strconv.Atoi(data[1]) //from
	if err != nil {
		log.Fatal("Unable to parse from", err)
	}

	m := MessageEvent{
		Event:    data[0],
		From:     from,
		Contents: strings.Join(data[2:], seperator),
	}

	return m
}
