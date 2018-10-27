package ai

import (
	"fmt"

	"github.com/chandler37/gobackgammon/brd"
)

// Returns a brd.Chooser. Until it's racing, it really, really hates open
// blots, even ones that are unhittable. (A blot is a point containing only one
// Checker.) TODO(chandler37): make it care less about unhittable ones but it
// should still care because those "unhittable" blots may make it harder to hit
// our opponent later.
//
// It detects if it's a race and plays differently then, delegating to
// PlayerRacer.
//
// It uses math/rand.Intn to choose when the heuristics leave more than one choice.
//
// If amountOfForesight == 0, returns a chooser that never performs Monte-Carlo
// simulations to help determine the best move. Otherwise, we will simulate
// some number of rolls to help us choose. If otherPlayer is not provided, we
// will use the result of MakePlayerConservative(0, nil). TODO(chandler37):
// Implement the simulations.
//
// TODO(chandler37): This does not avoid backgammons very well; see
// TestPlayerConservativeAvoidsBackgammons. That might be fine in a tournament
// depending on the score, but we have another TODO: TODO(chandler37):
// Implement tournament play.
func MakePlayerConservative(amountOfForesight uint64, otherPlayer brd.Chooser) brd.Chooser {
	if amountOfForesight < 1 {
		return playerConservative
	}
	if otherPlayer == nil {
		otherPlayer = MakePlayerConservative(0, nil)
	}
	panic("TODO(chandler37): implement me")
}

// A conservative Chooser with no foresight.
func playerConservative(choices []brd.Board) int {
	if debug {
		fmt.Printf("DBG(PlayerConservative): %d choices\n", len(choices))
	}
	if len(choices) == 1 {
		return 0
	}
	racing := true
	for _, choice := range choices {
		if !choice.Racing() {
			racing = false
		}
	}
	if racing {
		return PlayerRacer(choices)
	}
	nextRound := converter(choices)
	nextRound = minimizer(
		"minMyBlotLiability",
		nextRound,
		func(b *brd.Board) int64 {
			return int64(b.BlotLiability(b.Roller))
		})
	nextRound = minimizer(
		"minMyBlots",
		nextRound,
		func(b *brd.Board) (numBlots int64) {
			for _, p := range b.Pips[1:25] {
				if p[0] == b.Roller && p[1] == brd.NoChecker {
					numBlots++
				}
			}
			return
		})
	nextRound = maximizer(
		"maxOpponentPipCount",
		nextRound,
		func(b *brd.Board) int64 {
			return int64(b.PipCount(b.Roller.OtherColor()))
		})
	/*
	TODO(chandler37): if maxMyBlockedPoints is more important than maxPrimeSize it affects the following:
	White goes first.
	{W to play   41; !dbl; 1:WW 2: 3: 4: 5: 6:rrrrr 7: 8:rrr 9: 10: 11: 12:WWWWW 13:rrrrr 14: 15: 16: 17:WWW 18: 19:WWWWW 20: 21: 22: 23: 24:rr}
	{r to play 3333; !dbl; 1:WW 2: 3: 4: 5: 6:rrrrr 7: 8:rrr 9: 10: 11: 12:WWWW 13:rrrrr 14: 15: 16: 17:WWWW 18: 19:WWWWW 20: 21: 22: 23: 24:rr}
	What is the best move? making a 3-prime at 6,7,8? Or blocking more points?
	*/
	nextRound = maximizer(
		"maxMyBlockedPoints",
		nextRound,
		func(b *brd.Board) int64 {
			return int64(b.NumPointsBlocked(b.Roller))
		})
	nextRound = maximizer(
		"maxPrimeSize",
		nextRound,
		func(b *brd.Board) int64 {
			return int64(b.LengthOfMaxPrime(b.Roller))
		})
	nextRound = maximizer(
		"maxNumCheckersInMyHome",
		nextRound,
		func(b *brd.Board) int64 {
			return int64(b.NumCheckersHome(b.Roller))
		})
	answer := randomChoice(nextRound)
	if debug {
		fmt.Printf("DBG(PlayerConservative): Decided on %d from range [0, %d)\n", answer, len(choices))
	}
	return answer
}
