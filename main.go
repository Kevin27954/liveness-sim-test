package main

import (
	"bufio"
	"fmt"
	"log"
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

	numNodes := 2
	var wg sync.WaitGroup

	for i := 0; i < numNodes; i++ {
		var urlList []string
		for j := i + 1; j < numNodes; j++ {
			url := fmt.Sprintf("%s%s:%d/%s", protocol, domain, startingPort+j, endpoint)
			urlList = append(urlList, url)
		}

		connList := strings.Join(urlList, ",")
		fmt.Println(connList)

		wg.Add(1)

		go func() {
			defer wg.Done()

			color := i
			strPort := strconv.Itoa(startingPort + i)

			startNodeCmd := exec.Command("go", "run", "cmd/main.go", strPort, connList)
			fmt.Println("Ran: ", startNodeCmd.Args)

			logPipe, err := startNodeCmd.StdoutPipe()
			assert.NoError(err, "Unable to get Pipe")

			err = startNodeCmd.Start()
			// defer startNodeCmd.Wait()
			assert.NoError(err, "Unable to start CMD")

			reader := bufio.NewReader(logPipe)
			scanner := bufio.NewScanner(reader)

			for scanner.Scan() {
				log.Printf("\033[%dm%s\033[0m", color+31, scanner.Text())
			}
		}()
	}

	wg.Wait()
}
