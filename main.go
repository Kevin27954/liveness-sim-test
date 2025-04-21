package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	// "github.com/go-cmd/cmd"
	"github.com/Kevin27954/liveness-sim-test/assert"
)

func main() {

	protocol := "ws://"
	domain := "localhost"
	startingPort := 8000
	endpoint := "internal"

	numNodes := 3
	var wg sync.WaitGroup

	logFile, err := os.Create("logs.md")
	assert.NoError(err, "Error creating log file")
	var mutex sync.Mutex

	for i := 0; i < numNodes; i++ {
		var urlList []string
		for j := i + 1; j < numNodes; j++ {
			url := fmt.Sprintf("%s%s:%d/%s", protocol, domain, startingPort+j, endpoint)
			urlList = append(urlList, url)
		}

		connList := strings.Join(urlList, ",")
		fmt.Println(connList)

		wg.Add(1)

		go func(l *sync.Mutex) {
			defer wg.Done()

			color := i
			strPort := strconv.Itoa(startingPort + i)

			startNodeCmd := exec.Command("go", "run", "cmd/main.go", strPort, connList)
			fmt.Println("Ran: ", startNodeCmd.Args)

			logPipe, err := startNodeCmd.StdoutPipe()
			assert.NoError(err, "Unable to get Pipe")
			errPipe, err := startNodeCmd.StderrPipe()
			assert.NoError(err, "Unable to get Pipe")

			err = startNodeCmd.Start()
			// defer startNodeCmd.Wait()
			assert.NoError(err, "Unable to start CMD")

			scanner := bufio.NewScanner(logPipe)
			errScanner := bufio.NewScanner(errPipe)

			go func(l *sync.Mutex) {
				for errScanner.Scan() {
					text := errScanner.Text()
					log.Printf("\033[%dm%s\033[0m", color+31, text)
					l.Lock()
					logFile.Write([]byte(text + "\n"))
					l.Unlock()
				}
			}(l)

			for scanner.Scan() {
				text := scanner.Text()
				log.Printf("\033[%dm%s\033[0m", color+31, text)
				l.Lock()
				logFile.Write([]byte(text + "\n"))
				l.Unlock()
			}
		}(&mutex)
	}

	wg.Wait()
}
