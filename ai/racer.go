package ai

import (
	"github.com/chandler37/gobackgammon/brd"
)

// A brd.Chooser built for a race (a.k.a. a bear-off). It dislikes being
// backgammoned or gammmoned and prefers to get as many checkers off as
// possible. You would be silly to play this way if your opponent had a chance
// to hit you.
//
// TODO(chandler37): To avoid being gammoned we try to maxMyCheckersAtHome, but
// sometimes this might mean a move 7=>1 when that six would be better spent
// moving the farthest checker, say on the 17, closer to home. ai_test.go has a
// test case with a TODO demonstrating this.
func PlayerRacer(choices []*brd.Board) []brd.AnalyzedBoard {
	if len(choices) == 1 {
		return nil
	}
	nextRound := converter(choices)
	minimizer(
		"minProbabilityOfGettingBackgammoned",
		nextRound,
		probabilityOfGettingBackgammoned)
	maximizer(
		"maxMyCheckersBorneOff",
		nextRound,
		func(b *brd.Board) int64 {
			pip := brd.BorneOffRedPip
			if b.Roller == brd.White {
				pip = brd.BorneOffWhitePip
			}
			return int64(b.Pips[pip].NumCheckers())
		})
	// TODO(chandler37): Care about the pip count of not-at-home checkers. See
	// http://localhost:8000/game?s=qlYqKlWyUjIxJDKkioxgQQVhGcNZJnCWKZxlBmdBvGgM92O4CdyT4XBPQgXBIWSqVAsIAAD__w&t=
	maximizer(
		"maxMyCheckersAtHome",
		nextRound,
		func(b *brd.Board) int64 {
			return int64(b.NumCheckersHome(b.Roller))
		})
	minimizer(
		"minHowFarAwayMyFarthestIs",
		nextRound,
		func(b *brd.Board) int64 {
			return int64(b.PipCountOfFarthestChecker(b.Roller))
		})
	shuffle(nextRound)
	return nextRound
}

// TODO(chandler37): Fix the following bug: it thought there were two choices
// at the end:
// http://localhost:8000/game?s=qlYqKlWyUjIxxBpSRfCQMjSGhVWRESyMimBhBBECWVpkAnddONx14QjXmSnVAgIAAP__&t=
