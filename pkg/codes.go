package pkg

// Due to the standard lib also just comparing string for bytes. These codes are defaulted to byte strings.
// Unless there is a faster way, I don't see a need to change this.
const (
	CONSENSUS_YES = "0"
	CONSENSUS_NO  = "1"

	AM_LEADER = "5"

	APPEND_ENTRIES = "8"

	HEARTBEAT = "9"
	ELECTION  = "10"
	VOTE_YES  = "11"
	VOTE_NO   = "12"

	NEW_OP = "15"

	// Operation Codes Start After at 20

	NEW_MSG_ADD    = "100"
	NEW_MSG_DELETE = "101"

	SYNC_REQ_INIT   = "22"
	SYNC_REQ_ASK    = "23"
	SYNC_REQ_HAS    = "24"
	SYNC_REQ_NO_HAS = "25"
	SYNC_REQ_COMMIT = "26"
)
