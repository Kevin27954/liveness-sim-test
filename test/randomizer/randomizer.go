package randomizer

import (
	"math/rand"
)

type Randomizer struct {
	randGen *rand.Rand
}

func Init(seed int64) Randomizer {
	src := rand.New(rand.NewSource(seed))

	r := Randomizer{
		randGen: src,
	}

	return r
}

func (r *Randomizer) GetIntN(n int) int {
	return int(r.randGen.Intn(n))
}

func (r *Randomizer) GetIntRange(min int, max int) int {
	return min + int(r.randGen.Intn(max-min))
}

// From 0 -> 127
func (r *Randomizer) GetInt7() int {
	return int(r.randGen.Intn(127))
}

// From 0 -> 32767
func (r *Randomizer) GetInt15() int {
	return int(r.randGen.Intn(32767))
}

// From 0 -> 2147483647
func (r *Randomizer) GetInt31() int {
	return int(r.randGen.Intn(2147483647))
}
