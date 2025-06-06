package main

import (
	test "github.com/Kevin27954/liveness-sim-test/test"
	rand "github.com/Kevin27954/liveness-sim-test/test/randomizer"
	srv "github.com/Kevin27954/liveness-sim-test/test/server"
)

func startTest(nodes int) {
	randomizer := rand.Init(69)
	server := srv.Init(8000, nodes)
	simTest := test.Init(server, randomizer, nodes)

	simTest.StartTest()
}
