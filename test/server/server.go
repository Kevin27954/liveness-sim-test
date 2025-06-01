package server

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
)

type Server struct {
	StartingPort int
	NumNodes     int
}

func Init(startingPort int, numNodes int) Server {
	s := Server{
		StartingPort: startingPort,
		NumNodes:     numNodes,
	}

	return s
}

func (s *Server) Start() {
	for i := range s.NumNodes {
		var urlList []string
		for j := i + 1; j < s.NumNodes; j++ {
			url := strconv.Itoa(s.StartingPort + j)
			urlList = append(urlList, url)
		}

		connList := strings.Join(urlList, ",")
		fmt.Println(connList)

		go s.startServer(s.StartingPort+i, connList)
	}
}

// port is the port of the node that left.
func (s *Server) Rejoin(port int) {
	var urlList []string
	for j := range s.NumNodes {
		if j != port%s.StartingPort {
			url := strconv.Itoa(s.StartingPort + j)
			urlList = append(urlList, url)
		}
	}

	connList := strings.Join(urlList, ",")
	fmt.Println("REJOIN:", connList)

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

	startNodeCmd := exec.Command("go", "run", "cmd/main.go", strconv.Itoa(port), connList)
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

func (s *Server) GetPorts() []string {
	var urlList []string

	for i := range s.NumNodes {
		url := strconv.Itoa(s.StartingPort + i)
		urlList = append(urlList, url)

		fmt.Println(urlList)
	}

	return urlList
}

func (s *Server) NumLeaders() int {
	numLeader := 0

	for i := range s.NumNodes {
		url := fmt.Sprintf("http://localhost:%d/leader", s.StartingPort+i)
		res, err := http.Get(url)
		if err != nil {
			log.Println("Unable to get response")
			continue
		}
		body, err := io.ReadAll(res.Body)
		if err != nil {
			log.Println("Unable to read body")
			continue
		}
		if string(body) == "true" {
			numLeader += 1
		}
	}

	return numLeader
}
