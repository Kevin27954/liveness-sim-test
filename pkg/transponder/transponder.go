package transponder

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	p "github.com/Kevin27954/liveness-sim-test/pkg"
	"github.com/gorilla/websocket"
)

const (
	WRITEWAIT = 10 * time.Second
	SEPERATOR = "≡" // A Hambuger menu looking thing
)

type Transponder struct {
	connList []*websocket.Conn
	from     int // Corresponds to the RAFT Node it is reference to.

	onRecv func(message any)
}

func Init(from int, size int) Transponder {
	t := Transponder{connList: make([]*websocket.Conn, size), from: from}

	return t
}

func (t *Transponder) StartConns(portList string) {
	if portList == "" {
		return
	}

	ports := strings.Split(portList, ",")

	time.Sleep(5 * time.Second)
	for _, port := range ports {
		port, err := strconv.Atoi(port)
		connStr := fmt.Sprintf("ws://localhost:%d/internal/%d", port, t.from)

		conn, _, err := websocket.DefaultDialer.Dial(connStr, nil)
		if err != nil {
			log.Println("Unable to Connect to conn:", err)
		}

		log.Printf("Connected to %s", connStr)
		t.AddConn(conn, port%8000)
	}
}

func (t *Transponder) listen(conn *websocket.Conn) {
	for {
		if conn == nil {
			log.Println(t.from, "it was NILLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLL")
			break
		}
		_, msg, err := conn.ReadMessage()
		if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNoStatusReceived) {
			log.Println(t.from, "Unable to read message: ", err)
			for i, internalConn := range t.connList {
				if internalConn == conn {
					t.connList[i] = nil
				}
			}
			break
		} else if err != nil {
			log.Println("VERY IMPORTANT MESSAGE HEREREE EEE err:", err)
			continue
		}

		t.onRecv(p.Init(string(msg)))
	}
}

func (t *Transponder) Write(msg []byte) {
	for _, conn := range t.connList {
		if conn == nil {
			continue
		}

		conn.SetWriteDeadline(time.Now().Add(WRITEWAIT))
		err := conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Println("From: ", t.from, "Error", err)
			return
		}
	}
}

// The ID is known ahead of time (last digit of port number)
func (t *Transponder) WriteTo(id int, msg []byte) error {
	if id > len(t.connList) {
		log.Fatal("id was out of bounds")
	}

	conn := t.connList[id]
	if conn == nil {
		log.Println("I was nil")
		return fmt.Errorf("Conn is nil, It was shut downed")
	}

	err := conn.SetWriteDeadline(time.Now().Add(WRITEWAIT))
	err = conn.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		return err
	}

	return nil
}

func (t *Transponder) OnRecv(action func(message any)) {
	t.onRecv = action
}

func (t *Transponder) AddConn(conn *websocket.Conn, id int) {
	log.Println("Adding a new connection")
	t.connList[id] = conn

	go t.listen(conn)
}

func (t *Transponder) CreateMsg(msgs ...any) []byte {
	msgByte := []byte("")
	for i, msg := range msgs {
		msgByte = fmt.Append(msgByte, msg)
		if i != len(msgs)-1 {
			msgByte = fmt.Append(msgByte, SEPERATOR)
		}
	}

	return msgByte
}

func (t *Transponder) GetTotalConns() int {
	return len(t.connList)
}

func (t *Transponder) String() string {
	var sb strings.Builder

	sb.WriteString("  Connection List: [")
	for k := range t.connList {
		sb.WriteString("\n")
		if t.connList[k] != nil {
			sb.WriteString(fmt.Sprintf("    Connected to node%d: true", k))
		} else {
			sb.WriteString(fmt.Sprintf("    Connected to node%d: false", k))
		}
	}
	sb.WriteString("\n]\n")

	return sb.String()
}
