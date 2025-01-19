package main

import (
	// "fmt"
	"fmt"
	"log"
	"net/http"

	"strconv"
)

func createNode(num int) Node {
	name := "node" + strconv.Itoa(num)

	return Node{name: name, conn: nil}
}

func main() {
	addr := ":8000"
	var nodeList [5]Node

	for i := 0; i < 5; i++ {
		nodeList[i] = createNode(i)
	}

	for i := 0; i < 5; i++ {
		http.HandleFunc("/"+nodeList[i].name, nodeList[1].Start)
	}

	// Lists all the nodes currently on
	http.HandleFunc("/nodes", func(w http.ResponseWriter, r *http.Request) {
		nodes := ""
		for i := 0; i < 5; i++ {
			nodes += nodeList[i].name + " "
		}
		fmt.Fprintf(w, "%s", nodes)
	})

	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal("Unable to start server: ", err)
	}

}
