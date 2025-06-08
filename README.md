### Overview

This project was just a way for me to get to know more about simulation testing, coding it, and how it would fit in an application. Thus the main focus is not the RAFT implementation but rather the simulation testing, although it took more time to code that rather than the simulation.

**RAFT** was implemented using goroutines, channels, and gorilla websocket connection (cause I didn't know what I was doing and randomly picked). Each node is connected through websocket connection and heartbeat is sent every 3 seconds. The logs are identified using binary search method, resulting in faster searches.

**Simualtion Testing** is ran with only 2 components, **Clients** connected to each **Server** that is up, sending messages to it constantly until it reaches the maximum amount, given randomly upon `Init(num)`. **Server** will be crashing/disconnecting constantly at random interval and reconnecting after a random time. **Clients** are expected to reconnect to the server when its back up and sending messages again to the same server it was originally connected to.
The randomness are determined thorugh a seed for easier replication.



### Bugs Found Thus Far
*Strike through means fixed*

- ~Ungraceful exit of servers~
- ~Fautly Leader Elections~
- ~Missing Information, Leading to Error Parsing~
- ~Incorrect handling of messages before election~
- Occasional concurrency error to websocket
- Reconnected nodes does not SYNC

### Running this Projecet.

Requirements to run the project:
- Golang
- Docker

To run this project, simply have **Golang** installed, and in the terminal, type:

`go run . -n N -t` to run the simulation test for 1 instance
N is the number of instances of nodes you want to start up

`go run . -n N` to run the program
N is the number of instances of nodes you want to start up


### Outcome

The project is mostly done, other than fixing the bugs found through simulation testing, there are some minor improvements I want to make to the testing. One of the main things is the need to clean up the SQLite files and running multiple instances.
Although I could set it up such that it goes from 0..N (where n is the number of test), as the file name and port numbers, that is not something I like to see. Using `Docker` should solve both these issues, making it easier to clean up and start up the test in parallel.

### Takeaways

The main goal was to learn about **Simulation Testing**, but I also managed to learn about **RAFT**. **RAFT** itself is just an algorithm, how it's implemented doesn't matter, so long as it follows these core principals:
- One leader each term
- Logs are replicated
- Logs should be syncing (Heartbeat or RPC)

Next is Simulation Testing. There wasn't anything too complicated about it other than planning and understanding what it is I was testing. In this project, the core thing I wanted tested was also the core principals of **RAFT**. However, I believe the code I wrote to accomplish the test is subpar. The fact that a single test can only be done through a command and multiple test is only possible by running that same command in the cli multiple times is **TERRIBlE**.
I think the thing that I dislike is the reliance on using `os.Exec` to start up the servers. I could have probably coded it such that the servers started without using CLI and same for the clients. This could have been done by allowing the `Server` and `Client` struct to receive a few extra args if needed or shove all the logic into a `Server.Start(args)`/`Client.Start(args)` func - aka **encapsulation**. Although the other thing I was concern aboout was the use of SQL file and the need to clean it up as well as ports, probably would have been able to be done with `Docker` and it's SDK rather than using the CLI, which is cleaner.

---
So for this part, **Note to self**: `os.Exec` should be a last resort for running something. It's a pain in the rectum. Try to just give it the args in the code itself, like pass it to the `struct` itself or through a function call - though in my project, it owuld just be to the `struct` itself. Encapsulate the logic and make it available to be used everywhere. The `main.go` can just be a CLI parser or `env` reader that later runs the server and everything else. This way I can have a test file also that can do the same thing, but for test only.
---

Logging is pretty important. Having a structed logging system and color coordination made reading a lot easier. In the future of my projects, I think I might want some pre determined templates or a way of logging that I always implement and use. It should make debugging and other things much easier and more fun to look at, rather than just black and white symbols.

