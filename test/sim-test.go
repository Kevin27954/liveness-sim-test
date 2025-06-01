package simtest

import (
	"fmt"
	"log"
	"sync"
	"time"

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
}

func Init(server srv.Server, randomizer rand.Randomizer) SimTest {
	s := SimTest{server: server, client: make([]cli.Client, 0), randomizer: randomizer}

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

	for i, port := range s.server.GetPorts() {
		connUrl := fmt.Sprintf("ws://localhost:%s/ws", port)
		s.client[i].Connect(connUrl)
		go func(l *sync.Mutex) {
			msgs := s.randomizer.GetInt7()
			log.Println(msgs)
			s.client[i].Start(msgs)
			l.Lock()
			if len(s.client) == 1 {
				s.client = nil
			} else {
				s.client[i] = s.client[len(s.client)-1]
				s.client = s.client[:len(s.client)-1]
			}
			l.Unlock()
		}(&l)
	}

	time.Sleep(5 * time.Second)

	// Some coniditon I will think of later:
	// Ideas:
	// finish all things, e.g. all messages, all server nodes closes, etc.
	// for a time limit, e.g. 2-3 minute max.
	// for a random time.
	for {
		// TODO: Work on timing out multiple servers too

		s.TimeoutRandomServer()
		time.Sleep(25 * time.Second)
		if s.server.NumLeaders() > 1 {
			log.Fatal("There was more than 1 leader")
			break
		}

		if s.client == nil {
			break
		}
	}

	// CLIENT IDEA:
	// I'm just gonna have it connect to the number of servers that started up
	// If it gets disconencted, it will just attempt to reconnect.
	// I will wait for it to finish it's total messages sent, similar to how
	// 		a real user might do.
	// Once finish, that should mark the end of a single test.
	// During the entire time the clients are sending data,
	// 		the servers will be constantly brought down and having elections happening.
	// It is expected for each server to be in the same state, and potentially even the
	// 		same number of messages that the clients had send.

	// CLIENT CONT:
	// However, that means that the clients will be in their own goroutine.
	// The Server might also be in their own goroutine?
	// So does that mean it will be a syncWaitGroup? for a single test?
	// Or would it be more like the server is just continously running it's timeout
	//		but the moment the client finishes, the entire thing closes?
	// I feel like this setup makes sense, since:
	// - servers should also be shutting down in the real world. Otherwise it is just running
	//		normally.
	//		- this also test the election process for the server
	// - The clients is just gonna be sending messages only.
	//		- this test the log replication and state consistent perhaps?

	log.Println("Finished Test, Result: Pass")
}

func (s *SimTest) TimeoutRandomServer() {
	nodeNum := s.randomizer.GetIntN(s.server.NumNodes) + s.server.StartingPort
	s.server.CloseServer(nodeNum)
	log.Println("Timing out node:", nodeNum)
	// 10 ~ 20 seconds before reconnecting
	time.Sleep(time.Duration(s.randomizer.GetIntN(10)+10) + time.Second)
	s.server.Rejoin(nodeNum)
}
