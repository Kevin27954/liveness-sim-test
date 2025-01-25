package main

import (
	// "fmt"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Kevin27954/liveness-sim-test/assert"
)

func createNode(num int) Node {
	name := "node" + strconv.Itoa(num)
	db := Init(name)

	return Node{name: name, conn: nil, db: db}
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

	http.HandleFunc("/get/{node}", func(w http.ResponseWriter, r *http.Request) {
		node, err := strconv.Atoi(r.PathValue("node"))
		if err != nil {
			http.Error(w, "Not a number", http.StatusBadRequest)
			return
		}

		if node >= len(nodeList) {
			http.Error(w, "Greater than # of nodes", http.StatusBadGateway)
			return
		}

		messages, err := nodeList[node].GetMessages()
		if err != nil {
			http.Error(w, "Error getting messages", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, messages)
	})

	err := http.ListenAndServe(addr, nil)
	assert.NoError(err, "Unable to start server")
}
