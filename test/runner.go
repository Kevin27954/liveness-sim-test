package test

/*

Things to test (basically features of RAFT) :
- the database being consistent with one another (replication)
- there being a leader always (election)
- Applicaition correctly times out? (basic stuff)
- messages are all able to be sent

checkLogState() {
make sure the the logs have no same index as leader,
every machine has the same state
	- leader must have same previous state must be the exact same with child
follows are never more than the leader. (term and index)
	- their greatest index should be <= leader greatest index
}

checkElectionInfo() {
always only 1 leader
leaders are reelected upon crashes
leader is always the highest term (>=)
check interval / time of each node must not overlap.
}

checkFinal() {
database is consistent with each other?
term must be the same at the tend?
leader must be the same at the end
}

checkSync() {
it should check the interval is within the [minInterval, maxInterval].
	- this shold just be within the functions itself however I do it.
}

checkRejoin() {
the logs must be equal to leader during lost
}

*/

func main() {
}
