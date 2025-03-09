package node

// Due to the standard lib also just comparing string for bytes. These codes are defaulted to byte strings.
// Unless there is a faster way, I don't see a need to change this.
const (
	VOTE_NO  = "0"
	VOTE_YES = "1"

	ELECTION = "10"
)
