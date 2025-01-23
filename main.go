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

	fmt.Println("I ran here")

	db := Init("test")
	// res, err := db.conn.Exec("CREATE TABLE message (id INTEGER PRIMARY KEY, message TEXT);")
	// if err != nil {
	// 	log.Fatal("Err exec:", err)
	// }
	res, err := db.conn.Exec("INSERT INTO message (message) VALUES ('This is a test message 2')")
	row := db.conn.QueryRow("SELECT * FROM message")

	var message string
	var id int
	row.Scan(&id, &message)

	fmt.Println("Message:", message)

	log.Fatal("Res:", res)

	// Lists all the nodes currently on
	http.HandleFunc("/nodes", func(w http.ResponseWriter, r *http.Request) {
		nodes := ""
		for i := 0; i < 5; i++ {
			nodes += nodeList[i].name + " "
		}
		fmt.Fprintf(w, "%s", nodes)
	})

	err = http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal("Unable to start server: ", err)
	}

}
