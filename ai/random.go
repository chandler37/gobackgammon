package ai

import (
	"math/rand"

	"github.com/chandler37/gobackgammon/brd"
)

// Chooses at random.
func PlayerRandom(choices []brd.Board) int {
	return rand.Intn(len(choices))
}
