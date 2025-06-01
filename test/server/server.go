package server

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

type Server struct {
	startingPort int
	NumNodes     int
}

func Init(startingPort int, numNodes int) Server {
	s := Server{
		startingPort: startingPort,
		NumNodes:     numNodes,
	}

	return s
}

func (s *Server) Start() []string {
	var urlList []string

	for i := range s.NumNodes {
		for j := i + 1; j < s.NumNodes; j++ {
			url := string(s.startingPort + j)
			urlList = append(urlList, url)
		}

		connList := strings.Join(urlList, ",")
		fmt.Println(connList)

		go s.startServer(s.startingPort+i, connList)
	}

	return urlList
}

// port is the port of the node that left.
func (s *Server) Rejoin(port int) {
	var urlList []string
	for j := 0; j < s.NumNodes; j++ {
		if j != port%s.startingPort {
			url := string(s.startingPort + j)
			urlList = append(urlList, url)
		}
	}

	connList := strings.Join(urlList, ",")
	fmt.Println(connList)

	go s.startServer(port, connList)
}

func (s *Server) CloseServer(port int) {
	startNodeCmd := exec.Command("curl", fmt.Sprintf("http://localhost:%d/quit", port))
	err := startNodeCmd.Start()
	if err != nil {
		log.Fatal("Unable to start CMD to quit")
	}
}

func (s *Server) startServer(port int, connList string) {
	color := port

	startNodeCmd := exec.Command("go", "run", "cmd/main.go", string(port), connList)
	fmt.Println("Ran: ", startNodeCmd.Args)

	logPipe, err := startNodeCmd.StdoutPipe()
	if err != nil {
		log.Fatal("Unable to get Pipe")
	}
	errPipe, err := startNodeCmd.StderrPipe()
	if err != nil {
		log.Fatal("Unable to get Pipe")
	}

	err = startNodeCmd.Start()
	if err != nil {
		log.Fatal("Unable to start CMD")
	}

	scanner := bufio.NewScanner(logPipe)
	errScanner := bufio.NewScanner(errPipe)

	go func() {
		for errScanner.Scan() {
			text := errScanner.Text()
			log.Printf("\033[%dm%s\033[0m", color+31, text)

		}
	}()
	for scanner.Scan() {
		text := scanner.Text()
		log.Printf("\033[%dm%s\033[0m", color+31, text)
	}

}
