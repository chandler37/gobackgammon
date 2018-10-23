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

func BenchmarkPlayGameTimesTenConservative(b *testing.B) {
	rand.Seed(time.Now().UnixNano())
	numBoards, numRedWins, numWhiteWins, numBackgammons := 0, 0, 0, 0
	for n := 0; n < b.N; n++ {
		for nn := 0; nn < 10; nn++ { // without this loop you'll have to run 'go test' many times to get a good number because the variance is large
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
	}
	b.Logf(
		"numWhiteWins=%d numRedWins=%d numBackgammons=%d numBoards=%d\n",
		numWhiteWins, numRedWins, numBackgammons, numBoards)
}

func TestPlayGame(t *testing.T) {
	type example struct {
		seed         int64
		victor       brd.Checker
		stakes       int
		score        string
		numBoard     int
		lastLog      string
		chooser      func([]brd.Board) int
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
			43,
			"{r after playing 4444; !dbl; 1: 2: 3: 4: 5: 6: 7: 8: 9: 10: 11: 12: 13: 14: 15: 16: 17: 18: 19:WWWWW 20: 21: 22:WW 23: 24:, 8 W off, 15 r off, Score{Goal:0,W:0,r:1,Crawford on,inactive}}",
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
			48,
			"{r after playing   52; !dbl; 1: 2: 3: 4: 5: 6: 7: 8: 9: 10: 11: 12: 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:WWWW, 11 W off, 15 r off, Score{Goal:0,W:0,r:1,Crawford on,inactive}}",
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
		} else {
			for _, line := range log {
				t.Logf("exNum=%d: line is %v\n", exNum, line)
			}
			if l := log[len(log)-1]; l != ex.lastLog {
				t.Fatalf("ex=%v log[-1]=%v", ex, l)
			}
		}
	}
}

func prettyChoices(c []brd.Board) string {
	parts := make([]string, 0, len(c))
	for i, b := range c {
		parts = append(parts, fmt.Sprintf("%-3d %s", i, b.String()))
	}
	return strings.Join(parts, "\n")
}

func TestPlayerConservative(t *testing.T) {
	b := brd.New(true)
	b.Roller = Red
	b.Roll = brd.Roll{5, 1}
	choices := b.LegalContinuations()
	if choice := playerConservative(choices); choices[choice].String() != "{r after playing   51; !dbl; 1:WW 2: 3: 4: 5: 6:rrrrr 7: 8:rrrr 9: 10: 11: 12:WWWWW 13:rrrr 14: 15: 16: 17:WWW 18: 19:WWWWW 20: 21: 22: 23:r 24:r}" {
		t.Errorf("starting with\n%v\nchoice was %d (%v)\nfrom\n%v", b.String(), choice, choices[choice].String(), prettyChoices(choices))
	}
}

func TestPlayerConservative2(t *testing.T) {
	b := brd.New(true)
	b.Roller = Red
	b.Roll = brd.Roll{3, 1}
	choices := b.LegalContinuations()
	if choice := playerConservative(choices); choices[choice].String() != "{r after playing   31; !dbl; 1:WW 2: 3: 4: 5:rr 6:rrrr 7: 8:rr 9: 10: 11: 12:WWWWW 13:rrrrr 14: 15: 16: 17:WWW 18: 19:WWWWW 20: 21: 22: 23: 24:rr}" {
		t.Errorf("choice (starting from %v)\nwas %d (%v)\nfrom\n%v", b.String(), choice, choices[choice], prettyChoices(choices))
	}
}

func TestPlayerConservative3(t *testing.T) {
	b := brd.New(true)
	b.Roller = Red
	b.Roll = brd.Roll{5, 4}
	choices := b.LegalContinuations()
	if choice := playerConservative(choices); choices[choice].String() != "{r after playing   54; !dbl; 1:WW 2: 3: 4: 5: 6:rrrrr 7: 8:rrrr 9: 10: 11: 12:WWWWW 13:rrrr 14: 15: 16: 17:WWW 18: 19:WWWWW 20:r 21: 22: 23: 24:r}" {
		t.Errorf("choice (starting from %v)\nwas %d (%v)\nfrom\n%v", b.String(), choice, choices[choice], prettyChoices(choices))
	}
}

func TestPlayerConservativeAvoidsBackgammons(t *testing.T) {
	// {r to play   51; !dbl; 1:rrrrrr 2:rrrrrr 3:r 4: 5: 6: 7: 8: 9:r 10: 11: 12: 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23:WWWWWWW 24:r, 8 W off}
	b := brd.New(true)
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
	choices := b.LegalContinuations()
	// TODO(chandler37): Eliminate this mistake, except when it's not a mistake
	// (tournament play where a loss of Stakes vs. a loss of 3*Stakes is
	// identically bad as regards the tournament outcome):
	if choice := playerConservative(choices); choices[choice].String() != "{r after playing   51; !dbl; 1:rrrrrr 2:rrrrrr 3:rr 4: 5: 6: 7: 8: 9: 10: 11: 12: 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23:WWWWWWW 24:r, 8 W off}" {
		t.Errorf("choice (starting from %v)\nwas %d (%v)\nfrom\n%v", b.String(), choice, choices[choice], prettyChoices(choices))
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
		func(_ []brd.Board) int {
			panic("i will not be called")
		})
	if choice := chooser(choices); choices[choice].String() != "{r after playing   54; !dbl; 1:WW 2: 3: 4: 5: 6:rrrrr 7: 8:rrrr 9: 10: 11: 12:WWWWW 13:rrrr 14: 15: 16: 17:WWW 18: 19:WWWWW 20:r 21: 22: 23: 24:r}" {
		t.Errorf("choice (starting from %v)\nwas %d (%v)\nfrom\n%v", b.String(), choice, choices[choice], prettyChoices(choices))
	}
	chooser = MakePlayerConservative(0, nil)
	if choice := chooser(choices); choices[choice].String() != "{r after playing   54; !dbl; 1:WW 2: 3: 4: 5: 6:rrrrr 7: 8:rrrr 9: 10: 11: 12:WWWWW 13:rrrr 14: 15: 16: 17:WWW 18: 19:WWWWW 20:r 21: 22: 23: 24:r}" {
		t.Errorf("choice (starting from %v)\nwas %d (%v)\nfrom\n%v", b.String(), choice, choices[choice], prettyChoices(choices))
	}
}

func TestPlayerRandom(t *testing.T) {
	rand.Seed(3737373737373737)
	b := brd.New(true)
	choices := b.LegalContinuations()
	chooser := PlayerRandom
	if len(choices) != 7 {
		t.Fatalf("I'm testing that this can pick the last choice given so I need multiple choices: %d", len(choices))
	}
	if choice := chooser(choices); choice != len(choices)-1 {
		t.Errorf("choice (starting from %v)\nwas %d (%v)\nfrom\n%v", b.String(), choice, choices[choice], prettyChoices(choices))
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
			"{r after playing   51; !dbl; 1:rr 2:rrrrr 3:rrrr 4: 5: 6:rrr 7: 8: 9: 10: 11: 12: 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:WWWWWW, 9 W off, 1 r off}"},

		example{
			func(b *brd.Board) {
			},
			"{r after playing   63; !dbl; 1:WW 2: 3: 4: 5: 6:rrrrr 7: 8:rrr 9: 10: 11: 12:WWWWW 13:rrrrr 14: 15: 16: 17:WWW 18:r 19:WWWWW 20: 21:r 22: 23: 24:}"},

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
			"{r after playing   21; !dbl; 1: 2: 3: 4: 5: 6:rrrrrrrrrrrrrr 7: 8: 9: 10:r 11: 12: 13: 14: 15: 16: 17: 18: 19:WWWWWWWWWWWWWWW 20: 21: 22: 23: 24:}"},

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
			"{r after playing   61; !dbl; 1:r 2:rrrrrrrrrrrr 3: 4: 5: 6:r 7: 8: 9: 10: 11: 12: 13: 14: 15: 16: 17:r 18: 19:WWWWWWWWWWWWWWW 20: 21: 22: 23: 24:}",
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
			"{r after playing   61; !dbl; 1: 2:rrrrrrrrrrrr 3: 4: 5: 6:rr 7: 8: 9: 10:r 11: 12: 13: 14: 15: 16: 17: 18: 19:WWWWWWWWWWWWWWW 20: 21: 22: 23: 24:}",
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
			"{W after playing   61; !dbl; 1: 2: 3: 4: 5: 6:rrrrrrrrrrrrrrr 7: 8: 9: 10: 11: 12: 13: 14:W 15: 16: 17: 18: 19:WW 20: 21: 22: 23:WWWWWWWWWWWW 24:}",
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
		choice := PlayerRacer(choices)
		if x := choices[choice].String(); x != ex.Choice {
			t.Errorf("starting with\n%v\nchoice was %d (%v)\nfrom\n%v", b.String(), choice, x, prettyChoices(choices))
		}
	}
}
