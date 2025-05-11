package main

import (
	"log"
)

type TestA struct {
	fieldA string
}

func (t TestA) EditAndLog1(newS string) {
	t.fieldA = newS
	log.Println(t.fieldA)
}

func (t *TestA) EditAndLog2(newS string) {
	t.fieldA = newS
	log.Println(t.fieldA)
}

func main() {
	tA1 := TestA{fieldA: "stra"}
	tA2 := TestA{fieldA: "strb"}

	log.Println("Before edit")

	tA1.EditAndLog1("str2a")
	tA2.EditAndLog2("str2b")

	log.Println("After edit")

	log.Println(tA1.fieldA)
	log.Println(tA2.fieldA)
}
