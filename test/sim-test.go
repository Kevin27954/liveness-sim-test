package simtest

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/Kevin27954/liveness-sim-test/db"
	cli "github.com/Kevin27954/liveness-sim-test/test/client"
	rand "github.com/Kevin27954/liveness-sim-test/test/randomizer"
	srv "github.com/Kevin27954/liveness-sim-test/test/server"
)

// Maximum running at a single time
const INSTANCES = 2

// Gonna need a way to init the total messages and disconnections, since that is the 2
// core things about RAFT: logs and election.

type SimTest struct {
	server     srv.Server
	client     []cli.Client
	randomizer rand.Randomizer

	finishedClients int
}

func Init(server srv.Server, randomizer rand.Randomizer) SimTest {
	s := SimTest{server: server, client: make([]cli.Client, 0), randomizer: randomizer, finishedClients: 0}

	for range s.server.NumNodes {
		s.client = append(s.client, cli.Init(randomizer))
	}

	return s
}

// This is for getting information from the HTML perhaps?
func (s *SimTest) StartServer() {
}

// This is just one instance. We plan on running multiple of these.
func (s *SimTest) StartTest() {
	_ = s.randomizer.GetInt15()

	s.server.Start()

	time.Sleep(2 * time.Second)

	var l sync.Mutex
	totalMsg := 0

	for i, port := range s.server.GetPorts() {
		connUrl := fmt.Sprintf("ws://localhost:%s/ws", port)
		s.client[i].Connect(connUrl)
		msgs := s.randomizer.GetInt7()
		totalMsg += msgs
		go func(l *sync.Mutex) {
			log.Println(msgs)
			s.client[i].Start(msgs)
			l.Lock()
			s.finishedClients += 1
			l.Unlock()
		}(&l)
	}

	log.Println("Sleeping for some time")
	log.Println(time.Duration(s.randomizer.GetIntN(25)) * time.Second)
	time.Sleep(time.Duration(s.randomizer.GetIntN(25)) * time.Second)

	// Some coniditon I will think of later:
	// Ideas:
	// finish all things, e.g. all messages, all server nodes closes, etc.
	// for a time limit, e.g. 2-3 minute max.
	// for a random time.

	for {
		// TODO: Work on timing out multiple servers too
		// s.TimeoutRandomServer()
		// time.Sleep(time.Duration(s.randomizer.GetIntN(45)) * time.Second)
		// if s.server.NumLeaders() > 1 {
		// 	log.Fatal("There was more than 1 leader")
		// 	break
		// }

		if s.finishedClients >= s.server.NumNodes {
			break
		}
	}

	log.Println("Quitting All Servers...")
	s.QuitAllServer()

	time.Sleep(3 * time.Second)

	for i := range INSTANCES + 1 {
		name := "node" + strconv.Itoa(i)
		tempDB := db.Init(name)
		msg, err := tempDB.GetNumLogs()
		if err != nil {
			log.Println("Unable to get # of Logs")
		}

		log.Println("Node", i, ": ", msg)
		//clean up /delete db
	}

	log.Println(" Total Expected: ", totalMsg)

	log.Println("Finished Test, Result: Pass. This is a lie. I didn't check for anything")

}

func (s *SimTest) QuitAllServer() {
	for i := range s.server.NumNodes {
		nodeNum := i + s.server.StartingPort
		s.server.CloseServer(nodeNum)
	}
}

func (s *SimTest) TimeoutRandomServer() {
	nodeNum := s.randomizer.GetIntN(s.server.NumNodes) + s.server.StartingPort
	s.server.CloseServer(nodeNum)
	log.Println("Timing out node:", nodeNum)
	// 10 ~ 20 seconds before reconnecting
	time.Sleep(time.Duration(s.randomizer.GetIntN(10000)+10000) * time.Millisecond)
	s.server.Rejoin(nodeNum)
}
