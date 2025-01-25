package assert

import (
	"log"
	"os"
)

func runAssert(message string, data []any) {
	log.Print(message)
	for _, msg := range data {
		log.Printf("%s", msg)
	}

	os.Exit(1)
}

func NotNil(object any, message string, data ...any) {
	if object == nil {
		runAssert(message, append(data, "not null:", object))
	}
}

func Assert(truthy bool, got any, expected any, message string, data ...any) {
	if !truthy {
		runAssert(message, append(data, "got", got, "expected", expected))
	}
}

func NoError(err error, message string, data ...any) {
	if err != nil {
		runAssert(message, append(data, "error", err))
	}
}
