package ai

import (
	"math/rand"

	"github.com/chandler37/gobackgammon/brd"
)

type randomAnalysis struct{}

func (r *randomAnalysis) Summary() string {
	return "Picked at random from the set of distinct positions"
}

var theRandomAnalysis = randomAnalysis{}

// Chooses at random.
//
// This is not uniformly random if you treat moving the same checker 2 and then
// 3 as being different from moving it 3 and then 2. It is uniformly random in
// terms of final positions.
func PlayerRandom(choices []*brd.Board) []brd.AnalyzedBoard {
	if len(choices) < 2 {
		return nil
	}
	return []brd.AnalyzedBoard{brd.AnalyzedBoard{Board: choices[rand.Intn(len(choices))]}}
}
