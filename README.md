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

Next is Simulation Testing. There wasn't anything too complicated about it other than planning and understanding what it is I was testing. In this project, the core thing I wanted tested was also the core principals of **RAFT**. So it was pretty easier to do in. 

Logging is pretty important. Having a structed logging system and color coordination made reading a lot easier. In the future of my projects, I think I might want some pre determined templates or a way of logging that I always implement and use. It should make debugging and other things much easier and more fun to look at, rather than just black and white symbols.
