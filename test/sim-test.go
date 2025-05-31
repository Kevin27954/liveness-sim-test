package simtest

import (
	"log"

	cli "github.com/Kevin27954/liveness-sim-test/test/client"
	rand "github.com/Kevin27954/liveness-sim-test/test/randomizer"
	srv "github.com/Kevin27954/liveness-sim-test/test/server"
)

// Maximum running at a single time
const INSTANCES = 3

// Gonna need a way to init the total messages and disconnections, since that is the 2
// core things about RAFT: logs and election.

type SimTest struct {
	server     srv.Server
	client     cli.Client
	randomizer rand.Randomizer
}

func Init(server srv.Server, client cli.Client, randomizer rand.Randomizer) SimTest {
	s := SimTest{server: server, client: client, randomizer: randomizer}

	return s
}

// This is for getting information from the HTML perhaps?
func (s *SimTest) StartServer() {
}

// This is just one instance. We plan on running multiple of these.
func (s *SimTest) StartTest() {
	log.Println(s.randomizer.GetInt15())
}
