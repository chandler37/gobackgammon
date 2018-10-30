package ai

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/chandler37/gobackgammon/brd"
)

const (
	Red   = brd.Red
	White = brd.White
)

func BenchmarkPlayGameConservativeRandomlySeeded(b *testing.B) {
	rand.Seed(time.Now().UnixNano())
	helpBenchmarkPlayGameConservative(b)
}

func BenchmarkPlayGameConservativeReproduciblySeeded(b *testing.B) {
	rand.Seed(37)
	helpBenchmarkPlayGameConservative(b)
}

func helpBenchmarkPlayGameConservative(b *testing.B) {
	numBoards, numRedWins, numWhiteWins, numBackgammons := 0, 0, 0, 0
	for n := 0; n < b.N; n++ {
		logger := func(_ interface{}, b *brd.Board) {
			numBoards++
		}
		chooser := playerConservative
		victor, stakes, _ := brd.New(false).PlayGame(struct{}{}, chooser, logger, nil, nil)
		if victor == White {
			numWhiteWins++
		} else {
			numRedWins++
		}
		if stakes == 3 {
			numBackgammons++
		}
	}
	b.Logf(
		"numWhiteWins=%d numRedWins=%d numBackgammons=%d numBoards=%d\n",
		numWhiteWins, numRedWins, numBackgammons, numBoards)
}

func TestAiPlayGame(t *testing.T) {
	type example struct {
		seed         int64
		victor       brd.Checker
		stakes       int
		score        string
		numBoard     int
		lastLog      string
		chooser      brd.Chooser
		logger       func(interface{}, *brd.Board)
		offerDouble  func(*brd.Board) bool
		acceptDouble func(*brd.Board) bool
	}
	examples := [...]example{
		example{
			1337,
			Red,
			1,
			"Score{Goal:0,W:0,r:1,Crawford on,inactive}",
			45,
			"{r to play   33 after playing   33; !dbl; 1: 2: 3: 4: 5: 6: 7: 8: 9: 10: 11: 12: 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:W, 14 W off, 15 r off, Score{Goal:0,W:0,r:1,Crawford on,inactive}}",
			playerConservative,
			func(state interface{}, b *brd.Board) {
				if iv := b.Invalidity(brd.IgnoreRollValidity); iv != "" {
					t.Fatalf("invalidity=%v", iv)
				}
				slicePtr := state.(*[]string)
				*slicePtr = append(*slicePtr, b.String())
			},
			nil,
			nil},

		example{
			1338,
			Red,
			1,
			"Score{Goal:0,W:0,r:1,Crawford on,inactive}",
			50,
			"{r to play    3 after playing    5; !dbl; 1: 2: 3: 4: 5: 6: 7: 8: 9: 10: 11: 12: 13: 14: 15: 16: 17: 18: 19: 20: 21: 22:WWWWWW 23:WWW 24:WWWWW, 1 W off, 15 r off, Score{Goal:0,W:0,r:1,Crawford on,inactive}}",
			playerConservative,
			func(state interface{}, b *brd.Board) {
				if iv := b.Invalidity(brd.IgnoreRollValidity); iv != "" {
					t.Fatalf("invalidity=%v", iv)
				}
				slicePtr := state.(*[]string)
				*slicePtr = append(*slicePtr, b.String())
			},
			nil,
			nil},
	}

	for exNum, ex := range examples {
		log := []string{}
		rand.Seed(ex.seed)
		victor, stakes, score := brd.New(true).PlayGame(
			&log, ex.chooser, ex.logger, ex.offerDouble, ex.acceptDouble)
		if victor != ex.victor || stakes != ex.stakes || score.String() != ex.score {
			t.Errorf("exNum=%v ex=%v victor=%v stakes=%v score=%v(expected %v)", exNum, ex, victor, stakes, score, ex.score)
		}
		if len(log) != ex.numBoard {
			t.Errorf("exNum=%d ex=%v len(log)=%d", exNum, ex, len(log))
		}
		for _, line := range log {
			t.Logf("exNum=%d: line is %v\n", exNum, line)
		}
		if l := log[len(log)-1]; l != ex.lastLog {
			t.Fatalf("ex=%v log[-1]=%v", ex, l)
		}
	}
}

func prettyChoices(c []*brd.Board) string {
	parts := make([]string, 0, len(c))
	for i, b := range c {
		parts = append(parts, fmt.Sprintf("%-3d %s", i, b.String()))
	}
	return strings.Join(parts, "\n")
}

func prettyAnalyzedChoices(c []brd.AnalyzedBoard) string {
	parts := make([]string, 0, len(c))
	for i, ab := range c {
		parts = append(parts, fmt.Sprintf("%-3d %v", i, ab))
	}
	return strings.Join(parts, "\n")
}

func TestPlayerConservative(t *testing.T) {
	type example struct {
		Initializer func(*brd.Board)
		Choice      string
		Analyzed    []string
	}
	examples := [...]example{
		example{
			func(b *brd.Board) {
				b.Roller = Red
				b.Roll = brd.Roll{5, 1}
			},
			"{r after playing   51; !dbl; 1:WW 2: 3: 4: 5: 6:rrrrr 7: 8:rrrr 9: 10: 11: 12:WWWWW 13:rrrr 14: 15: 16: 17:WWW 18: 19:WWWWW 20: 21: 22: 23:r 24:r}",
			nil},
		example{
			func(b *brd.Board) {
				b.Roller = Red
				b.Roll = brd.Roll{3, 1}
			},
			"{r after playing   31; !dbl; 1:WW 2: 3: 4: 5:rr 6:rrrr 7: 8:rr 9: 10: 11: 12:WWWWW 13:rrrrr 14: 15: 16: 17:WWW 18: 19:WWWWW 20: 21: 22: 23: 24:rr}",
			nil},
		example{
			func(b *brd.Board) {
				b.Roller = Red
				b.Roll = brd.Roll{5, 4}
			},
			"{r after playing   54; !dbl; 1:WW 2: 3: 4: 5: 6:rrrrr 7: 8:rrrr 9: 10: 11: 12:WWWWW 13:rrrr 14: 15: 16: 17:WWW 18: 19:WWWWW 20:r 21: 22: 23: 24:r}",
			nil},
		example{
			func(b *brd.Board) {
				// avoids backgammons: {r to play   51; !dbl; 1:rrrrrr 2:rrrrrr 3:r 4: 5: 6: 7: 8: 9:r 10: 11: 12: 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23:WWWWWWW 24:r, 8 W off}
				b.Roller = Red
				b.Roll = brd.Roll{5, 1}
				b.Pips = brd.Points28{}
				b.Pips[1].Reset(6, Red)
				b.Pips[2].Reset(6, Red)
				b.Pips[3].Reset(1, Red)
				b.Pips[9].Reset(1, Red)
				b.Pips[24].Reset(1, Red)
				b.Pips[23].Reset(7, White)
				b.Pips[brd.BorneOffWhitePip].Reset(8, White)
				// TODO(chandler37): Eliminate this mistake (moving the 3
				// instead of the 19), except when it's not a mistake
				// (tournament play where a loss of Stakes vs. a loss of
				// 3*Stakes is identically bad as regards the tournament
				// outcome)
			},
			"{r after playing   51; !dbl; 1:rrrrrr 2:rrrrrrr 3: 4: 5: 6: 7: 8: 9:r 10: 11: 12: 13: 14: 15: 16: 17: 18: 19:r 20: 21: 22: 23:WWWWWWW 24:, 8 W off}",
			nil},
		example{
			func(b *brd.Board) {
				// {r to play   43; !dbl; 1:rrrrrr 2:WW 3:rrr 4:rrr 5: 6:rr 7: 8: 9: 10: 11: 12: 13: 14: 15: 16:W 17:WWWWWW 18: 19:WWWWWW 20: 21:r 22: 23: 24:}
				//
				// which is interesting because it has but one legal
				// continuation since there's only one way to use the entire
				// roll.
				b.Roller = Red
				b.Roll = brd.Roll{4, 3}
				b.Pips = brd.Points28{}
				b.Pips[1].Reset(6, Red)
				b.Pips[2].Reset(2, White)
				b.Pips[3].Reset(3, Red)
				b.Pips[4].Reset(3, Red)
				b.Pips[6].Reset(2, Red)
				b.Pips[16].Reset(1, White)
				b.Pips[17].Reset(6, White)
				b.Pips[19].Reset(6, White)
				b.Pips[21].Reset(1, Red)
			},
			"{r after playing   34; !dbl; 1:rrrrrr 2:WW 3:rrr 4:rrr 5: 6:rr 7: 8: 9: 10: 11: 12: 13: 14:r 15: 16:W 17:WWWWWW 18: 19:WWWWWW 20: 21: 22: 23: 24:}",
			nil},
		example{
			func(b *brd.Board) {
				// {r to play   41; !dbl; 1: 2:rrr 3:rrr 4:r 5:rr 6:rr 7: 8: 9: 10: 11: 12: 13:rrr 14: 15: 16: 17:r 18: 19: 20:WW 21: 22:WWWW 23:WWW 24:WW, 4 W off}

				// TODO(chandler37): Add a new heuristic so that {r after
				// playing 41; !dbl; 1: 2:rrr 3:rrr 4:r 5:rr 6:rr 7: 8: 9: 10:
				// 11: 12:r 13:rrr 14: 15: 16: 17: 18: 19: 20:WW 21: 22:WWWW
				// 23:WWW 24:WW, 4 W off} is the chosen move because the pip
				// count outside of home is most reduced.
				b.Roller = Red
				b.Roll = brd.Roll{4, 1}
				b.Pips = brd.Points28{}
				b.Pips[2].Reset(3, Red)
				b.Pips[3].Reset(3, Red)
				b.Pips[4].Reset(1, Red)
				b.Pips[5].Reset(2, Red)
				b.Pips[6].Reset(2, Red)
				b.Pips[13].Reset(3, Red)
				b.Pips[17].Reset(1, Red)
				b.Pips[20].Reset(2, White)
				b.Pips[22].Reset(4, White)
				b.Pips[23].Reset(3, White)
				b.Pips[24].Reset(2, White)
				b.Pips[brd.BorneOffWhitePip].Reset(4, White)
			},
			"{r after playing   41; !dbl; 1: 2:rrr 3:rrr 4:r 5:rrr 6:r 7: 8: 9: 10: 11: 12: 13:rrrr 14: 15: 16: 17: 18: 19: 20:WW 21: 22:WWWW 23:WWW 24:WW, 4 W off}",
			[]string{
				"{r after playing   41; !dbl; 1: 2:rrr 3:rrr 4:r 5:rrr 6:r 7: 8: 9: 10: 11: 12: 13:rrrr 14: 15: 16: 17: 18: 19: 20:WW 21: 22:WWWW 23:WWW 24:WW, 4 W off} (wasn't ruled out by heuristics. Details: maxMyCheckersAtHome=11 maxMyCheckersBorneOff=0 minHowFarAwayMyFarthestIs=13 minProbabilityOfGettingBackgammoned=0 randomizer=8734748956570804315)",
				"{r after playing   41; !dbl; 1: 2:rrr 3:rrr 4:rr 5:r 6:rr 7: 8: 9: 10: 11: 12: 13:rrrr 14: 15: 16: 17: 18: 19: 20:WW 21: 22:WWWW 23:WWW 24:WW, 4 W off} (Ruled out by randomizer (2112854853711373400))",
				"{r after playing   41; !dbl; 1: 2:rrrr 3:rr 4:r 5:rr 6:rr 7: 8: 9: 10: 11: 12: 13:rrrr 14: 15: 16: 17: 18: 19: 20:WW 21: 22:WWWW 23:WWW 24:WW, 4 W off} (Ruled out by randomizer (3543339188040204842))",
				"{r after playing   41; !dbl; 1: 2:rrr 3:rrr 4:r 5:rr 6:rr 7: 8: 9: 10: 11: 12:r 13:rrr 14: 15: 16: 17: 18: 19: 20:WW 21: 22:WWWW 23:WWW 24:WW, 4 W off} (Ruled out by randomizer (4753501743317724670))",
				"{r after playing   41; !dbl; 1:r 2:rr 3:rrr 4:r 5:rr 6:rr 7: 8: 9: 10: 11: 12: 13:rrrr 14: 15: 16: 17: 18: 19: 20:WW 21: 22:WWWW 23:WWW 24:WW, 4 W off} (Ruled out by randomizer (5999264223995105160))",
				"{r after playing   41; !dbl; 1: 2:rrr 3:rrrr 4: 5:rr 6:rr 7: 8: 9: 10: 11: 12: 13:rrrr 14: 15: 16: 17: 18: 19: 20:WW 21: 22:WWWW 23:WWW 24:WW, 4 W off} (Ruled out by randomizer (8530525039392881481))",
				"{r after playing   41; !dbl; 1:r 2:rrr 3:rrr 4:r 5:r 6:rr 7: 8: 9: 10: 11: 12: 13:rrr 14: 15: 16:r 17: 18: 19: 20:WW 21: 22:WWWW 23:WWW 24:WW, 4 W off} (Ruled out by minHowFarAwayMyFarthestIs (16))",
				"{r after playing   41; !dbl; 1: 2:rrrr 3:rrr 4:r 5:rr 6:r 7: 8: 9: 10: 11: 12: 13:rrr 14: 15: 16:r 17: 18: 19: 20:WW 21: 22:WWWW 23:WWW 24:WW, 4 W off} (Ruled out by minHowFarAwayMyFarthestIs (16))",
				"{r after playing   41; !dbl; 1: 2:rrr 3:rrr 4:r 5:rr 6:rr 7: 8: 9:r 10: 11: 12: 13:rr 14: 15: 16:r 17: 18: 19: 20:WW 21: 22:WWWW 23:WWW 24:WW, 4 W off} (Ruled out by minHowFarAwayMyFarthestIs (16))",
				"{r after playing   41; !dbl; 1:rr 2:rr 3:rrr 4:r 5:r 6:rr 7: 8: 9: 10: 11: 12: 13:rrr 14: 15: 16: 17:r 18: 19: 20:WW 21: 22:WWWW 23:WWW 24:WW, 4 W off} (Ruled out by minHowFarAwayMyFarthestIs (17))",
				"{r after playing   41; !dbl; 1:r 2:rrrr 3:rr 4:r 5:r 6:rr 7: 8: 9: 10: 11: 12: 13:rrr 14: 15: 16: 17:r 18: 19: 20:WW 21: 22:WWWW 23:WWW 24:WW, 4 W off} (Ruled out by minHowFarAwayMyFarthestIs (17))",
				"{r after playing   41; !dbl; 1:r 2:rrr 3:rrrr 4: 5:r 6:rr 7: 8: 9: 10: 11: 12: 13:rrr 14: 15: 16: 17:r 18: 19: 20:WW 21: 22:WWWW 23:WWW 24:WW, 4 W off} (Ruled out by minHowFarAwayMyFarthestIs (17))",
				"{r after playing   41; !dbl; 1:r 2:rrr 3:rrr 4:rr 5: 6:rr 7: 8: 9: 10: 11: 12: 13:rrr 14: 15: 16: 17:r 18: 19: 20:WW 21: 22:WWWW 23:WWW 24:WW, 4 W off} (Ruled out by minHowFarAwayMyFarthestIs (17))",
				"{r after playing   41; !dbl; 1:r 2:rrr 3:rrr 4:r 5:rr 6:r 7: 8: 9: 10: 11: 12: 13:rrr 14: 15: 16: 17:r 18: 19: 20:WW 21: 22:WWWW 23:WWW 24:WW, 4 W off} (Ruled out by minHowFarAwayMyFarthestIs (17))",
				"{r after playing   41; !dbl; 1:r 2:rrr 3:rrr 4:r 5:r 6:rr 7: 8: 9: 10: 11: 12:r 13:rr 14: 15: 16: 17:r 18: 19: 20:WW 21: 22:WWWW 23:WWW 24:WW, 4 W off} (Ruled out by minHowFarAwayMyFarthestIs (17))",
				"{r after playing   41; !dbl; 1: 2:rrrrr 3:rr 4:r 5:rr 6:r 7: 8: 9: 10: 11: 12: 13:rrr 14: 15: 16: 17:r 18: 19: 20:WW 21: 22:WWWW 23:WWW 24:WW, 4 W off} (Ruled out by minHowFarAwayMyFarthestIs (17))",
				"{r after playing   41; !dbl; 1: 2:rrrr 3:rrrr 4: 5:rr 6:r 7: 8: 9: 10: 11: 12: 13:rrr 14: 15: 16: 17:r 18: 19: 20:WW 21: 22:WWWW 23:WWW 24:WW, 4 W off} (Ruled out by minHowFarAwayMyFarthestIs (17))",
				"{r after playing   41; !dbl; 1: 2:rrrr 3:rrr 4:rr 5:r 6:r 7: 8: 9: 10: 11: 12: 13:rrr 14: 15: 16: 17:r 18: 19: 20:WW 21: 22:WWWW 23:WWW 24:WW, 4 W off} (Ruled out by minHowFarAwayMyFarthestIs (17))",
				"{r after playing   41; !dbl; 1: 2:rrrr 3:rrr 4:r 5:rrr 6: 7: 8: 9: 10: 11: 12: 13:rrr 14: 15: 16: 17:r 18: 19: 20:WW 21: 22:WWWW 23:WWW 24:WW, 4 W off} (Ruled out by minHowFarAwayMyFarthestIs (17))",
				"{r after playing   41; !dbl; 1: 2:rrrr 3:rrr 4:r 5:rr 6:r 7: 8: 9: 10: 11: 12:r 13:rr 14: 15: 16: 17:r 18: 19: 20:WW 21: 22:WWWW 23:WWW 24:WW, 4 W off} (Ruled out by minHowFarAwayMyFarthestIs (17))",
				"{r after playing   41; !dbl; 1:r 2:rr 3:rrr 4:r 5:rr 6:rr 7: 8: 9:r 10: 11: 12: 13:rr 14: 15: 16: 17:r 18: 19: 20:WW 21: 22:WWWW 23:WWW 24:WW, 4 W off} (Ruled out by minHowFarAwayMyFarthestIs (17))",
				"{r after playing   41; !dbl; 1: 2:rrrr 3:rr 4:r 5:rr 6:rr 7: 8: 9:r 10: 11: 12: 13:rr 14: 15: 16: 17:r 18: 19: 20:WW 21: 22:WWWW 23:WWW 24:WW, 4 W off} (Ruled out by minHowFarAwayMyFarthestIs (17))",
				"{r after playing   41; !dbl; 1: 2:rrr 3:rrrr 4: 5:rr 6:rr 7: 8: 9:r 10: 11: 12: 13:rr 14: 15: 16: 17:r 18: 19: 20:WW 21: 22:WWWW 23:WWW 24:WW, 4 W off} (Ruled out by minHowFarAwayMyFarthestIs (17))",
				"{r after playing   41; !dbl; 1: 2:rrr 3:rrr 4:rr 5:r 6:rr 7: 8: 9:r 10: 11: 12: 13:rr 14: 15: 16: 17:r 18: 19: 20:WW 21: 22:WWWW 23:WWW 24:WW, 4 W off} (Ruled out by minHowFarAwayMyFarthestIs (17))",
				"{r after playing   41; !dbl; 1: 2:rrr 3:rrr 4:r 5:rrr 6:r 7: 8: 9:r 10: 11: 12: 13:rr 14: 15: 16: 17:r 18: 19: 20:WW 21: 22:WWWW 23:WWW 24:WW, 4 W off} (Ruled out by minHowFarAwayMyFarthestIs (17))",
				"{r after playing   41; !dbl; 1: 2:rrr 3:rrr 4:r 5:rr 6:rr 7: 8:r 9: 10: 11: 12: 13:rr 14: 15: 16: 17:r 18: 19: 20:WW 21: 22:WWWW 23:WWW 24:WW, 4 W off} (Ruled out by minHowFarAwayMyFarthestIs (17))",
				"{r after playing   41; !dbl; 1: 2:rrr 3:rrr 4:r 5:rr 6:rr 7: 8: 9:r 10: 11: 12:r 13:r 14: 15: 16: 17:r 18: 19: 20:WW 21: 22:WWWW 23:WWW 24:WW, 4 W off} (Ruled out by minHowFarAwayMyFarthestIs (17))",
			}},
	}
	for exNum, ex := range examples {
		rand.Seed(37)
		b := brd.New(true)
		ex.Initializer(b)
		if iv := b.Invalidity(brd.EnforceRollValidity); iv != "" {
			t.Fatalf("iv for %d: %v", exNum, iv)
		}
		choices := b.LegalContinuations()
		if len(choices) == 0 {
			t.Fatalf("b is %v", *b)
		}
		pchoices := playerConservative(choices)
		if choice := pchoices[0]; choice.Board.String() != ex.Choice {
			t.Errorf("exNum=%d starting with\n%v\nchoice was (%v)\nfrom\n%v", exNum, b.String(), choice, prettyAnalyzedChoices(pchoices))
		}
		if len(ex.Analyzed) > 0 {
			actual := []string{}
			for _, pc := range pchoices {
				actual = append(actual, pc.String())
			}
			if len(actual) != len(ex.Analyzed) {
				t.Errorf("exNum=%d len(actual)=%d len(ex.Analyzed)=%d actual=\n%v", exNum, len(actual), len(ex.Analyzed), prettyAnalyzedChoices(pchoices))
			} else {
				for i, a := range actual {
					if x := ex.Analyzed[i]; a != x {
						t.Errorf("exNum=%d i=%d a=%v x=%v", exNum, i, a, x)
					}
				}
			}
		}
	}
}

func TestMakePlayerConservative(t *testing.T) {
	StartDebugging()
	StopDebugging() // a net noop
	b := brd.New(true)
	b.Roller = Red
	b.Roll = brd.Roll{5, 4}
	choices := b.LegalContinuations()
	chooser := MakePlayerConservative(
		0,
		func(_ []*brd.Board) []brd.AnalyzedBoard {
			panic("i will not be called")
		})
	analyzedChoices := chooser(choices)
	if cs := analyzedChoices[0].Board.String(); cs != "{r after playing   54; !dbl; 1:WW 2: 3: 4: 5: 6:rrrrr 7: 8:rrrr 9: 10: 11: 12:WWWWW 13:rrrr 14: 15: 16: 17:WWW 18: 19:WWWWW 20:r 21: 22: 23: 24:r}" {
		t.Errorf("choice (starting from %v)\nwas %v\nfrom\n%v", b.String(), cs, prettyChoices(choices))
	}
	chooser = MakePlayerConservative(0, nil)
	analyzedChoices = chooser(choices)
	if cs := analyzedChoices[0].Board.String(); cs != "{r after playing   54; !dbl; 1:WW 2: 3: 4: 5: 6:rrrrr 7: 8:rrrr 9: 10: 11: 12:WWWWW 13:rrrr 14: 15: 16: 17:WWW 18: 19:WWWWW 20:r 21: 22: 23: 24:r}" {
		t.Errorf("choice (starting from %v)\nwas %v\nfrom\n%v", b.String(), cs, prettyChoices(choices))
	}
}

func TestPlayerRandom(t *testing.T) {
	rand.Seed(3737373737373737)
	b := brd.New(true)
	choices := b.LegalContinuations()
	if len(choices) != 7 {
		t.Fatalf("I'm testing that this can pick the last choice given so I need multiple choices: %d", len(choices))
	}
	analyzedChoices := PlayerRandom(choices)
	if len(analyzedChoices) != 1 {
		t.Errorf("Why would it bother?")
	}
	if analyzedChoices[0].Board != choices[len(choices)-1] {
		t.Errorf("choice (starting from %v)\nwas %v\nfrom\n%v", b.String(), analyzedChoices[0].Board, prettyChoices(choices))
	}
	if analyzedChoices[0].Analysis != nil {
		t.Errorf("What analysis is valuable for a randomizer?")
	}
}

func TestPlayerRacer(t *testing.T) {
	type example struct {
		Initializer func(*brd.Board)
		Choice      string
	}
	examples := [...]example{
		example{
			func(b *brd.Board) {
				b.Roller = Red
				b.Roll = brd.Roll{5, 1}
				b.Pips = brd.Points28{}
				b.Pips[1].Reset(2, Red)
				b.Pips[2].Reset(5, Red)
				b.Pips[3].Reset(4, Red)
				b.Pips[6].Reset(4, Red)
				b.Pips[24].Reset(6, White)
				b.Pips[brd.BorneOffWhitePip].Reset(9, White)
			},
			"{r after playing   51; !dbl; 1:rr 2:rrrrr 3:rrrr 4: 5: 6:rrr 7: 8: 9: 10: 11: 12: 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:WWWWWW, 9 W off, 1 r off} (wasn't ruled out by heuristics. Details: maxMyCheckersAtHome=14 maxMyCheckersBorneOff=1 minHowFarAwayMyFarthestIs=6 minProbabilityOfGettingBackgammoned=0 randomizer=1926012586526624009)"},

		example{
			func(b *brd.Board) {
			},
			"{r after playing   63; !dbl; 1:WW 2: 3: 4: 5: 6:rrrrr 7: 8:rrr 9: 10: 11: 12:WWWWW 13:rrrrr 14: 15: 16: 17:WWW 18:r 19:WWWWW 20: 21:r 22: 23: 24:} (wasn't ruled out by heuristics. Details: maxMyCheckersAtHome=5 maxMyCheckersBorneOff=0 minHowFarAwayMyFarthestIs=21 minProbabilityOfGettingBackgammoned=3 randomizer=1926012586526624009)"},

		example{
			func(b *brd.Board) {
				b.Roller = Red
				b.Pips = brd.Points28{}
				b.Roll = brd.Roll{2, 1}
				b.Pips[19].Reset(15, White)
				b.Pips[6].Reset(13, Red)
				b.Pips[8].Reset(1, Red)
				b.Pips[11].Reset(1, Red)
			},
			"{r after playing   21; !dbl; 1: 2: 3: 4: 5: 6:rrrrrrrrrrrrrr 7: 8: 9: 10:r 11: 12: 13: 14: 15: 16: 17: 18: 19:WWWWWWWWWWWWWWW 20: 21: 22: 23: 24:} (wasn't ruled out by heuristics. Details: maxMyCheckersAtHome=14 maxMyCheckersBorneOff=0 minHowFarAwayMyFarthestIs=10 minProbabilityOfGettingBackgammoned=0 randomizer=1926012586526624009)"},

		example{
			func(b *brd.Board) {
				b.Roller = Red
				b.Pips = brd.Points28{}
				b.Roll = brd.Roll{6, 1}
				b.Pips[19].Reset(15, White)
				b.Pips[17].Reset(1, Red)
				b.Pips[7].Reset(2, Red)
				b.Pips[2].Reset(12, Red)
			},
			"{r after playing   61; !dbl; 1:r 2:rrrrrrrrrrrr 3: 4: 5: 6:r 7: 8: 9: 10: 11: 12: 13: 14: 15: 16: 17:r 18: 19:WWWWWWWWWWWWWWW 20: 21: 22: 23: 24:} (wasn't ruled out by heuristics. Details: maxMyCheckersAtHome=14 maxMyCheckersBorneOff=0 minHowFarAwayMyFarthestIs=17 minProbabilityOfGettingBackgammoned=0 randomizer=1926012586526624009)",
			// TODO(chandler37): Make PlayerRacer prefer this move:
			//
			// "{r after playing   61; !dbl; 1: 2:rrrrrrrrrrrr 3: 4: 5: 6:r 7:r 8: 9: 10: 11:r 12: 13: 14: 15: 16: 17: 18: 19:WWWWWWWWWWWWWWW 20: 21: 22: 23: 24:}"},
		},

		example{
			func(b *brd.Board) {
				b.Roller = Red
				b.Pips = brd.Points28{}
				b.Roll = brd.Roll{6, 1}
				b.Pips[19].Reset(15, White)
				b.Pips[17].Reset(1, Red)
				b.Pips[6].Reset(2, Red)
				b.Pips[2].Reset(12, Red)
			},
			"{r after playing   61; !dbl; 1: 2:rrrrrrrrrrrr 3: 4: 5: 6:rr 7: 8: 9: 10:r 11: 12: 13: 14: 15: 16: 17: 18: 19:WWWWWWWWWWWWWWW 20: 21: 22: 23: 24:} (wasn't ruled out by heuristics. Details: maxMyCheckersAtHome=14 maxMyCheckersBorneOff=0 minHowFarAwayMyFarthestIs=10 minProbabilityOfGettingBackgammoned=0 randomizer=1926012586526624009)",
		},

		example{
			func(b *brd.Board) {
				b.Roller = White
				b.Pips = brd.Points28{}
				b.Roll = brd.Roll{6, 1}
				b.Pips[6].Reset(15, Red)
				b.Pips[7].Reset(1, White)
				b.Pips[19].Reset(2, White)
				b.Pips[23].Reset(12, White)
			},
			"{W after playing   61; !dbl; 1: 2: 3: 4: 5: 6:rrrrrrrrrrrrrrr 7: 8: 9: 10: 11: 12: 13: 14:W 15: 16: 17: 18: 19:WW 20: 21: 22: 23:WWWWWWWWWWWW 24:} (wasn't ruled out by heuristics. Details: maxMyCheckersAtHome=14 maxMyCheckersBorneOff=0 minHowFarAwayMyFarthestIs=11 minProbabilityOfGettingBackgammoned=0 randomizer=1926012586526624009)",
		},
	}
	for exNum, ex := range examples {
		rand.Seed(42)
		b := brd.New(true)
		ex.Initializer(b)
		if iv := b.Invalidity(brd.EnforceRollValidity); iv != "" {
			t.Fatalf("exNum=%d invalidity:%v", exNum, iv)
		}
		choices := b.LegalContinuations()
		analyzedChoices := PlayerRacer(choices)
		if x := analyzedChoices[0].String(); x != ex.Choice {
			t.Errorf("exNum=%d starting with\n%v\nchoice was %v\nnot %v\nfrom\n%v", exNum, b.String(), analyzedChoices[0], ex.Choice, prettyChoices(choices))
		}
	}
}

func TestMinmaximizerAndConverter(t *testing.T) {
	rand.Seed(37)
	b := brd.New(true)
	firstThree := b.LegalContinuations()[:3]
	choices := converter(firstThree)
	if len(choices) != 3 {
		t.Fatalf("hmmm")
	}
	strs := []string{}
	for _, b := range firstThree {
		strs = append(strs, b.String())
	}
	t.Logf("strs is %v", strs)
	if x := choices[0].Board.String(); x != strs[0] {
		t.Fatalf("LegalContinuations changed or converter broke: %v", x)
	}
	scores := []int64{39, 37, 39}
	i := -1
	maximizer("foo", choices, func(b *brd.Board) int64 { i++; return scores[i] })
	expectedAnalyedBoards := []string{
		fmt.Sprintf("%s (wasn't ruled out by heuristics. Details: foo=39)", strs[0]),
		fmt.Sprintf("%s (wasn't ruled out by heuristics. Details: foo=39)", strs[2]),
		fmt.Sprintf("%s (Ruled out by foo (37))", strs[1])}
	if len(expectedAnalyedBoards) != len(choices) {
		t.Fatalf("hmmm2")
	}
	for i, c := range choices {
		if x := c.String(); x != expectedAnalyedBoards[i] {
			t.Fatalf("%d: it is %v", i, x)
		}
	}

	// Another round
	maximizer("bar", choices, func(b *brd.Board) int64 { return -100 })
	expectedAnalyedBoards = []string{
		fmt.Sprintf("%s (wasn't ruled out by heuristics. Details: bar=-100 foo=39)", strs[0]),
		fmt.Sprintf("%s (wasn't ruled out by heuristics. Details: bar=-100 foo=39)", strs[2]),
		fmt.Sprintf("%s (Ruled out by foo (37))", strs[1])}
	if len(expectedAnalyedBoards) != len(choices) {
		t.Fatalf("hmmm2")
	}
	for i, c := range choices {
		if x := c.String(); x != expectedAnalyedBoards[i] {
			t.Fatalf("%d: it is %v", i, x)
		}
	}

	// Another round with minimizer
	scores = []int64{40, 37}
	i = -1
	minimizer("baz", choices, func(b *brd.Board) int64 { i++; return scores[i] })
	expectedAnalyedBoards = []string{
		fmt.Sprintf("%s (wasn't ruled out by heuristics. Details: bar=-100 baz=37 foo=39)", strs[2]),
		fmt.Sprintf("%s (Ruled out by baz (40))", strs[0]),
		fmt.Sprintf("%s (Ruled out by foo (37))", strs[1])}
	if len(expectedAnalyedBoards) != len(choices) {
		t.Fatalf("hmmm2")
	}
	for i, c := range choices {
		if x := c.String(); x != expectedAnalyedBoards[i] {
			t.Fatalf("%d: it is %v", i, x)
		}
	}
}
