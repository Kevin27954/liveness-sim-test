package node

// Due to the standard lib also just comparing string for bytes. These codes are defaulted to byte strings.
// Unless there is a faster way, I don't see a need to change this.
const (
	CONSENSUS_YES = "0"
	CONSENSUS_NO  = "1"

	AM_LEADER = "5"

	HEARTBEAT = "9"
	ELECTION  = "10"
	VOTE_YES  = "11"
	VOTE_NO   = "12"

	// Operation Codes Start After at 20

	NEW_MSG_ADD    = "20"
	NEW_MSG_DELETE = "21"

	SYNC_REQ_ASK    = "22"
	SYNC_REQ_HAS    = "23"
	SYNC_REQ_NO_HAS = "24"
	SYNC_REQ_COMMIT = "25"
)
