package ai

import (
	"fmt"

	"github.com/chandler37/gobackgammon/brd"
)

// DLC
func MakePlayerAggressive(amountOfForesight uint64, otherPlayer brd.Chooser) brd.Chooser {
	if amountOfForesight < 1 {
		return playerAggressive
	}
	if otherPlayer == nil {
		otherPlayer = MakePlayerAggressive(0, nil)
	}
	panic("TODO(chandler37): implement me")
}

// An aggressive Chooser with no foresight (0-ply).
func playerAggressive(choices []*brd.Board) []brd.AnalyzedBoard {
	if debug {
		fmt.Printf("DBG(PlayerAggressive): %d choices\n", len(choices))
	}
	if len(choices) == 1 {
		return []brd.AnalyzedBoard{brd.AnalyzedBoard{Board: choices[0]}}
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
	maximizer(
		"maxOpponentPipCount",
		nextRound,
		func(b *brd.Board) int64 {
			return int64(b.PipCount(b.Roller.OtherColor()))
		})
	minimizer(
		"minMyBlotLiability",
		nextRound,
		func(b *brd.Board) int64 {
			return int64(b.BlotLiability(b.Roller, false))
		})
	minimizer(
		"minProbabilityOfGettingBackgammoned",
		nextRound,
		probabilityOfGettingBackgammoned)
	minimizer(
		"minMyBlots",
		nextRound,
		func(b *brd.Board) (numBlots int64) {
			for _, p := range b.Pips[1:25] {
				if p.Num(b.Roller) == 1 {
					numBlots++
				}
			}
			return
		})
	/*
		TODO(chandler37): if maxMyBlockedPoints is more important than maxPrimeSize it affects the following:
		White goes first.
		{W to play   41; !dbl; 1:WW 2: 3: 4: 5: 6:rrrrr 7: 8:rrr 9: 10: 11: 12:WWWWW 13:rrrrr 14: 15: 16: 17:WWW 18: 19:WWWWW 20: 21: 22: 23: 24:rr}
		{r to play 3333; !dbl; 1:WW 2: 3: 4: 5: 6:rrrrr 7: 8:rrr 9: 10: 11: 12:WWWW 13:rrrrr 14: 15: 16: 17:WWWW 18: 19:WWWWW 20: 21: 22: 23: 24:rr}
		What is the best move? making a 3-prime at 6,7,8? Or blocking more points?
	*/
	maximizer(
		"maxMyBlockedPoints",
		nextRound,
		func(b *brd.Board) int64 {
			return int64(b.NumPointsBlocked(b.Roller))
		})
	maximizer(
		"maxPrimeSize",
		nextRound,
		func(b *brd.Board) int64 {
			return int64(b.LengthOfMaxPrime(b.Roller))
		})
	maximizer(
		"maxNumCheckersInMyHome",
		nextRound,
		func(b *brd.Board) int64 {
			return int64(b.NumCheckersHome(b.Roller))
		})
	minimizer(
		"minMyBlotLiabilityIncludingUnhittable",
		nextRound,
		func(b *brd.Board) int64 {
			return int64(b.BlotLiability(b.Roller, true))
		})
	shuffle(nextRound)
	return nextRound
}
