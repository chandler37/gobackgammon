package brd

import (
	"fmt"
	"math/rand"
	"runtime"
	"strings"
	"testing"
	"time"
	"unsafe"
)

func BenchmarkPlayGameTimesTen(b *testing.B) {
	rand.Seed(time.Now().UnixNano())
	numBoards, numRedWins, numWhiteWins, numBackgammons := 0, 0, 0, 0
	for n := 0; n < b.N; n++ {
		for nn := 0; nn < 10; nn++ { // without this loop you'll have to run 'go test' many times to get a good number because the variance is large
			logger := func(_ interface{}, b *Board) {
				numBoards++
			}
			chooser := func(s []*Board) []AnalyzedBoard {
				return []AnalyzedBoard{AnalyzedBoard{Board: s[rand.Intn(len(s))]}}
			}
			victor, stakes, _ := New(false).PlayGame(struct{}{}, chooser, logger, nil, nil)
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

func BenchmarkRollNew(b *testing.B) {
	tmp := Roll{}
	r := Roll{}
	for n := 0; n < b.N; n++ {
		r.New(&tmp)
	}
}

func BenchmarkLegalContinuationsWith96Possibilities(b *testing.B) {
	for n := 0; n < b.N; n++ {
		board := New(false)
		board.Roller = White
		board.Roll = Roll{6, 6, 6, 6}
		board.Pips = Points28{}
		board.Pips[7].Reset(3, White)
		board.Pips[6].Reset(4, White)
		board.Pips[5].Reset(4, White)
		board.Pips[4].Reset(4, White)
		board.Pips[BarRedPip].Reset(1, Red)
		board.Pips[BorneOffRedPip].Reset(14, Red)
		if len(board.LegalContinuations()) != 96 {
			panic("not 96")
		}
	}
}

func BenchmarkComingOffTheBar(b *testing.B) {
	for n := 0; n < b.N; n++ {
		board := New(false)
		board.Roller = White
		board.Roll = Roll{6, 6, 6, 6}
		board.Pips[19].Reset(0, White)
		board.Pips[BarWhitePip].Reset(5, White)
		board.Pips[6].Reset(1, Red)
		board.Pips[BarRedPip].Reset(4, Red)
		nextBoards := board.continuationsOffTheBar()
		if len(nextBoards) != 1 {
			panic(fmt.Sprintf("nextBoards is %v", nextBoards))
		}
		if x := nextBoards[0].Pips[6].String(); x != "WWWW" {
			panic(fmt.Sprintf("x is %v", x))
		}
	}
}

// This test case may not work on your GOOS/GOARCH. It may fail with a
// different Go version. Plus it's slow.
func TestRandomnessOfRolling(t *testing.T) {
	rand.Seed(43)
	tmp := Roll{1, 2, 3, 4}
	r := Roll{}
	occurrences := [...]int{0, 0, 0, 0, 0, 0}
	numDoubles, numTotal := 0, 1000000
	for n := 0; n < numTotal; n++ {
		r.New(&tmp)
		if len(tmp.Dice()) != 0 {
			t.Fatalf("should clear tmp")
		}
		occurrences[r[0]-1]++
		occurrences[r[1]-1]++
		if r[0] == r[1] {
			numDoubles++
		}
	}
	if occurrences != [...]int{333728, 333929, 333215, 333422, 332983, 332723} {
		t.Errorf("actual occurrences: %v", occurrences)
	}
	if numDoubles != 165578 {
		t.Errorf("numDoubles=%d", numDoubles)
	}
}

func TestNewAndStringerAndInvalidity(t *testing.T) {
	rand.Seed(42)
	b := New(true)
	if bs := fmt.Sprintf("%v", b); bs != "{r to play   63; !dbl; 1:WW 2: 3: 4: 5: 6:rrrrr 7: 8:rrr 9: 10: 11: 12:WWWWW 13:rrrrr 14: 15: 16: 17:WWW 18: 19:WWWWW 20: 21: 22: 23: 24:rr}" {
		t.Errorf("Bad initial board %v", bs)
	}
	b.Pips[BarWhitePip].Reset(1, White)
	if i := b.Invalidity(EnforceRollValidity); i != "16 White checkers found, not 15" {
		t.Errorf("should be invalid though: %v", i)
	}
	if b.Pips[1] != NewPoint(2, White) {
		t.Errorf("b.Pips[1] is %v", b.Pips[1])
	}
	if b.Pips[1] == NewPoint(0, White) || b.Pips[1] == NewPoint(1, White) || b.Pips[1] == NewPoint(2, Red) {
		t.Errorf("b.Pips[1] should not equal that")
	}
	b.Pips[1].Reset(1, White)
	if i := b.Invalidity(EnforceRollValidity); i != "" {
		t.Errorf("should be valid: %v", i)
	}
	if bs := fmt.Sprintf("%v", b); bs != "{r to play   63; !dbl; 1:W 2: 3: 4: 5: 6:rrrrr 7: 8:rrr 9: 10: 11: 12:WWWWW 13:rrrrr 14: 15: 16: 17:WWW 18: 19:WWWWW 20: 21: 22: 23: 24:rr, W on bar}" {
		t.Errorf("Bad b.String() %v", bs)
	}
	b.Pips[BarRedPip].Reset(2, Red)
	b.Pips[24].Reset(0, Red)
	b.Pips[BorneOffWhitePip].Reset(2, White)
	q := b.Pips[12].NumCheckers()
	if q != 5 {
		t.Errorf("%v != 5", q)
	}
	b.Pips[12].Reset(3, White)
	b.Pips[BorneOffRedPip].Reset(3, Red)
	b.Pips[13].Reset(2, Red)
	if i := b.Invalidity(EnforceRollValidity); i != "" {
		t.Errorf("validity %v", i)
	}
	if bs := fmt.Sprintf("%v", b); bs != "{r to play   63; !dbl; 1:W 2: 3: 4: 5: 6:rrrrr 7: 8:rrr 9: 10: 11: 12:WWW 13:rr 14: 15: 16: 17:WWW 18: 19:WWWWW 20: 21: 22: 23: 24:, W on bar, rr on bar, 2 W off, 3 r off}" {
		t.Errorf("Bad b.String() %v", bs)
	}
	b.Stakes = 64
	b.WhiteCanDouble = false
	if bs := fmt.Sprintf("%v", b); bs != "{r to play   63; Stakes: 64, W canNOT dbl, r can dbl; 1:W 2: 3: 4: 5: 6:rrrrr 7: 8:rrr 9: 10: 11: 12:WWW 13:rr 14: 15: 16: 17:WWW 18: 19:WWWWW 20: 21: 22: 23: 24:, W on bar, rr on bar, 2 W off, 3 r off}" {
		t.Errorf("Bad b.String() %v", bs)
	}

	{
		b := New(false)
		b.Roller = White
		b.Roll = Roll{6, 2}
		// TODO(chandler37): Make the AIs play differently when Goal is 1 or
		// 2. They have heuristics to avoid backgammons and gammons.
		b.MatchScore = Score{Goal: 1}
		assertValidity(b, t)
		if s := b.String(); s != "{W to play   62; !dbl; 1:WW 2: 3: 4: 5: 6:rrrrr 7: 8:rrr 9: 10: 11: 12:WWWWW 13:rrrrr 14: 15: 16: 17:WWW 18: 19:WWWWW 20: 21: 22: 23: 24:rr, Score{Goal:1,W:0,r:0,Crawford on,inactive}}" {
			t.Errorf("got %v", s)
		}
		b.MatchScore = Score{Goal: 1, WhiteScore: 2, NoCrawfordRule: true}
		assertValidity(b, t)
		if s := b.String(); s != "{W to play   62; !dbl; 1:WW 2: 3: 4: 5: 6:rrrrr 7: 8:rrr 9: 10: 11: 12:WWWWW 13:rrrrr 14: 15: 16: 17:WWW 18: 19:WWWWW 20: 21: 22: 23: 24:rr, Score{Goal:1,W:2,r:0,Crawford off}}" {
			t.Errorf("got %v", s)
		}
		b.MatchScore = Score{Goal: 1, WhiteScore: 2, AlreadyPlayedCrawfordGame: true}
		assertValidity(b, t)
		if s := b.String(); s != "{W to play   62; !dbl; 1:WW 2: 3: 4: 5: 6:rrrrr 7: 8:rrr 9: 10: 11: 12:WWWWW 13:rrrrr 14: 15: 16: 17:WWW 18: 19:WWWWW 20: 21: 22: 23: 24:rr, Score{Goal:1,W:2,r:0,Crawford on,dormant}}" {
			t.Errorf("got %v", s)
		}
	}
}

func prettyCandidates(candidates []*Board) string {
	parts := make([]string, 0, len(candidates))
	for _, c := range candidates {
		parts = append(parts, c.String())
	}
	return fmt.Sprintf("[\n%v\n]", strings.Join(parts, "\n"))
}

func TestLegalContinuations(t *testing.T) {
	type example struct {
		Initializer      func(*Board) // is passed a board from New()
		InitializerCheck string
		continuations    []string
	}
	examples := [...]example{
		// "When bearing off, a beginner player may sometimes arrive at a
		// position where it appears as if he has to move both the numbers of
		// his roll inside his board without taking a checker off, but that
		// depends on the position – if a player has four checkers left to bear
		// off, one on each of his 1, 2, 5 and 6 points, and the player rolls a
		// 4 and a 3, no checker comes off, the player must move the 4 and 3
		// using the checkers on his 5 and 6 point.
		//
		// "However, if a player has, for example, two checkers on his 1 point,
		// two on his 3 point and one on his 6 point, and rolls a 5-2, he does
		// not have to play   the 5 from the 6 point to the 1 point and then the
		// 2 from his 3 point to his 1 point. Instead, he may first play the 2,
		// from his 6 point to his 4 point, and then bear the checker off from
		// his 4 point using the 5 of the dice roll."
		example{
			func(board *Board) {
				board.Roller = Red
				board.Roll = Roll{4, 3}
				board.Pips = Points28{}
				board.Pips[24].Reset(1, White)
				board.Pips[1].Reset(1, Red)
				board.Pips[2].Reset(1, Red)
				board.Pips[5].Reset(1, Red)
				board.Pips[6].Reset(1, Red)
				board.Pips[BorneOffRedPip].Reset(11, Red)
				board.Pips[BorneOffWhitePip].Reset(14, White)
			},
			"{r to play   43; !dbl; 1:r 2:r 3: 4: 5:r 6:r 7: 8: 9: 10: 11: 12: 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:W, 14 W off, 11 r off}",
			[]string{
				"{r after playing   43; !dbl; 1:rr 2:r 3:r 4: 5: 6: 7: 8: 9: 10: 11: 12: 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:W, 14 W off, 11 r off}",
				"{r after playing   43; !dbl; 1:r 2:rrr 3: 4: 5: 6: 7: 8: 9: 10: 11: 12: 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:W, 14 W off, 11 r off}",
			}},
		example{ // a continuation of the above example
			func(board *Board) {
				board.Roller = Red
				board.Roll = Roll{5, 2}
				board.Pips = Points28{}
				board.Pips[24].Reset(1, White)
				board.Pips[1].Reset(2, Red)
				board.Pips[3].Reset(2, Red)
				board.Pips[6].Reset(1, Red)
				board.Pips[BorneOffRedPip].Reset(10, Red)
				board.Pips[BorneOffWhitePip].Reset(14, White)
			},
			"{r to play   52; !dbl; 1:rr 2: 3:rr 4: 5: 6:r 7: 8: 9: 10: 11: 12: 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:W, 14 W off, 10 r off}",
			[]string{
				"{r after playing   52; !dbl; 1:rrrr 2: 3:r 4: 5: 6: 7: 8: 9: 10: 11: 12: 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:W, 14 W off, 10 r off}",
				"{r after playing   25; !dbl; 1:rr 2: 3:rr 4: 5: 6: 7: 8: 9: 10: 11: 12: 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:W, 14 W off, 11 r off}",
			}},

		// white has one on the bar and rolls a 1,6 but is blocked from using
		// the six to come in and must come in on the 1. all of white's
		// checkers are at home and the 7 point is blocked by red so make sure
		// only the 1 is used, wasting the six, instead of naively using the
		// six to bear off a white checker
		example{
			func(board *Board) {
				board.Roller = White
				board.Roll = Roll{6, 1}
				board.Pips[BarWhitePip].Reset(1, White)
				board.Pips[20].Reset(5, White)
				board.Pips[21].Reset(4, White)
				board.Pips[17].Reset(0, White)
				board.Pips[12].Reset(0, White)
				board.Pips[1].Reset(0, White)
				board.Pips[7].Reset(2, Red)
				board.Pips[6].Reset(3, Red)
			},
			"{W to play   61; !dbl; 1: 2: 3: 4: 5: 6:rrr 7:rr 8:rrr 9: 10: 11: 12: 13:rrrrr 14: 15: 16: 17: 18: 19:WWWWW 20:WWWWW 21:WWWW 22: 23: 24:rr, W on bar}",
			[]string{"{W to play    6 after playing    1; !dbl; 1:W 2: 3: 4: 5: 6:rrr 7:rr 8:rrr 9: 10: 11: 12: 13:rrrrr 14: 15: 16: 17: 18: 19:WWWWW 20:WWWWW 21:WWWW 22: 23: 24:rr}"}},

		example{
			func(board *Board) {
				board.Roller = White
				board.Roll = Roll{6, 5}
			},
			"{W to play   65; !dbl; 1:WW 2: 3: 4: 5: 6:rrrrr 7: 8:rrr 9: 10: 11: 12:WWWWW 13:rrrrr 14: 15: 16: 17:WWW 18: 19:WWWWW 20: 21: 22: 23: 24:rr}",
			[]string{
				"{W after playing   65; !dbl; 1:W 2: 3: 4: 5: 6:rrrrr 7: 8:rrr 9: 10: 11: 12:WWWWWW 13:rrrrr 14: 15: 16: 17:WWW 18: 19:WWWWW 20: 21: 22: 23: 24:rr}", // 1 to 7, 7 to 12
				"{W after playing   65; !dbl; 1:W 2: 3: 4: 5: 6:rrrrr 7:W 8:rrr 9: 10: 11: 12:WWWW 13:rrrrr 14: 15: 16: 17:WWWW 18: 19:WWWWW 20: 21: 22: 23: 24:rr}", // 1 to 7, 12-17
				"{W after playing   65; !dbl; 1:W 2: 3: 4: 5: 6:rrrrr 7:W 8:rrr 9: 10: 11: 12:WWWWW 13:rrrrr 14: 15: 16: 17:WW 18: 19:WWWWW 20: 21: 22:W 23: 24:rr}", // 1 to 7, 17-22
				"{W after playing   65; !dbl; 1:WW 2: 3: 4: 5: 6:rrrrr 7: 8:rrr 9: 10: 11: 12:WWW 13:rrrrr 14: 15: 16: 17:WWWW 18:W 19:WWWWW 20: 21: 22: 23: 24:rr}", // 12-18, 12-17
				"{W after playing   65; !dbl; 1:WW 2: 3: 4: 5: 6:rrrrr 7: 8:rrr 9: 10: 11: 12:WWWW 13:rrrrr 14: 15: 16: 17:WW 18:W 19:WWWWW 20: 21: 22:W 23: 24:rr}", // 12-18, 17-22
				"{W after playing   65; !dbl; 1:WW 2: 3: 4: 5: 6:rrrrr 7: 8:rrr 9: 10: 11: 12:WWWW 13:rrrrr 14: 15: 16: 17:WWW 18: 19:WWWWW 20: 21: 22: 23:W 24:rr}", // 12-17, 17-23
				"{W after playing   65; !dbl; 1:WW 2: 3: 4: 5: 6:rrrrr 7: 8:rrr 9: 10: 11: 12:WWWWW 13:rrrrr 14: 15: 16: 17:W 18: 19:WWWWW 20: 21: 22:W 23:W 24:rr}", // 17-23, 17-22
			}},

		example{
			func(board *Board) {
				board.Roller = White
				board.Roll = Roll{6, 6, 6, 6}
				board.Pips = Points28{}
				board.Pips[BarWhitePip].Reset(3, White)
				board.Pips[BorneOffWhitePip].Reset(12, White)
				board.Pips[12].Reset(2, Red)
				board.Pips[BarRedPip].Reset(13, Red)
			},
			"{W to play 6666; !dbl; 1: 2: 3: 4: 5: 6: 7: 8: 9: 10: 11: 12:rr 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, WWW on bar, rrrrrrrrrrrrr on bar, 12 W off}",
			[]string{"{W to play    6 after playing  666; !dbl; 1: 2: 3: 4: 5: 6:WWW 7: 8: 9: 10: 11: 12:rr 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, rrrrrrrrrrrrr on bar, 12 W off}"}},

		example{
			func(board *Board) {
				board.Roller = White
				board.Roll = Roll{6, 6, 6, 6}
				board.Pips = Points28{}
				board.Pips[BarWhitePip].Reset(2, White)
				board.Pips[BorneOffWhitePip].Reset(13, White)
				board.Pips[12].Reset(2, Red)
				board.Pips[BarRedPip].Reset(13, Red)
			},
			"{W to play 6666; !dbl; 1: 2: 3: 4: 5: 6: 7: 8: 9: 10: 11: 12:rr 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, WW on bar, rrrrrrrrrrrrr on bar, 13 W off}",
			[]string{"{W to play   66 after playing   66; !dbl; 1: 2: 3: 4: 5: 6:WW 7: 8: 9: 10: 11: 12:rr 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, rrrrrrrrrrrrr on bar, 13 W off}"}},

		example{
			func(board *Board) {
				board.Roller = White
				board.Roll = Roll{6, 6, 6, 6}
				board.Pips = Points28{}
				board.Pips[BarWhitePip].Reset(1, White)
				board.Pips[BorneOffWhitePip].Reset(14, White)
				board.Pips[12].Reset(2, Red)
				board.Pips[BarRedPip].Reset(13, Red)
			},
			"{W to play 6666; !dbl; 1: 2: 3: 4: 5: 6: 7: 8: 9: 10: 11: 12:rr 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, W on bar, rrrrrrrrrrrrr on bar, 14 W off}",
			[]string{"{W to play  666 after playing    6; !dbl; 1: 2: 3: 4: 5: 6:W 7: 8: 9: 10: 11: 12:rr 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, rrrrrrrrrrrrr on bar, 14 W off}"}},

		example{ // a mirror
			func(board *Board) {
				board.Roller = Red
				board.Roll = Roll{6, 6, 6, 6}
				board.Pips = Points28{}
				board.Pips[BarRedPip].Reset(2, Red)
				board.Pips[BorneOffRedPip].Reset(13, Red)
				board.Pips[13].Reset(2, White)
				board.Pips[BarWhitePip].Reset(13, White)
			},
			"{r to play 6666; !dbl; 1: 2: 3: 4: 5: 6: 7: 8: 9: 10: 11: 12: 13:WW 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, WWWWWWWWWWWWW on bar, rr on bar, 13 r off}",
			[]string{"{r to play   66 after playing   66; !dbl; 1: 2: 3: 4: 5: 6: 7: 8: 9: 10: 11: 12: 13:WW 14: 15: 16: 17: 18: 19:rr 20: 21: 22: 23: 24:, WWWWWWWWWWWWW on bar, 13 r off}"}},

		example{
			func(board *Board) {
				board.Roller = Red
				board.Roll = Roll{6, 6, 6, 6}
				board.Pips[1].Subtract()
				board.Pips[12].Add(White)
			},
			"{r to play 6666; !dbl; 1:W 2: 3: 4: 5: 6:rrrrr 7: 8:rrr 9: 10: 11: 12:WWWWWW 13:rrrrr 14: 15: 16: 17:WWW 18: 19:WWWWW 20: 21: 22: 23: 24:rr}", // 1 to 7, 7 to 12
			[]string{
				// 8,8,8,*:
				"{r after playing 6666; !dbl; 1:W 2:rrr 3: 4: 5: 6:rrrrr 7:r 8: 9: 10: 11: 12:WWWWWW 13:rrrr 14: 15: 16: 17:WWW 18: 19:WWWWW 20: 21: 22: 23: 24:rr}",
				"{r after playing 6666; !dbl; 1:W 2:rrr 3: 4: 5: 6:rrrrr 7: 8: 9: 10: 11: 12:WWWWWW 13:rrrrr 14: 15: 16: 17:WWW 18:r 19:WWWWW 20: 21: 22: 23: 24:r}",
				// 8,8,*,*:
				"{r after playing 6666; !dbl; 1:r 2:rr 3: 4: 5: 6:rrrrr 7: 8:r 9: 10: 11: 12:WWWWWW 13:rrrr 14: 15: 16: 17:WWW 18: 19:WWWWW 20: 21: 22: 23: 24:rr, W on bar}",
				"{r after playing 6666; !dbl; 1:W 2:rr 3: 4: 5: 6:rrrrr 7:rr 8:r 9: 10: 11: 12:WWWWWW 13:rrr 14: 15: 16: 17:WWW 18: 19:WWWWW 20: 21: 22: 23: 24:rr}",
				"{r after playing 6666; !dbl; 1:W 2:rr 3: 4: 5: 6:rrrrr 7:r 8:r 9: 10: 11: 12:WWWWWW 13:rrrr 14: 15: 16: 17:WWW 18:r 19:WWWWW 20: 21: 22: 23: 24:r}",
				"{r after playing 6666; !dbl; 1:W 2:rr 3: 4: 5: 6:rrrrr 7: 8:r 9: 10: 11: 12:WWWWWW 13:rrrrr 14: 15: 16: 17:WWW 18:rr 19:WWWWW 20: 21: 22: 23: 24:}",
				// 8,*,*,*:
				//     8,13,7,13:
				"{r after playing 6666; !dbl; 1:r 2:r 3: 4: 5: 6:rrrrr 7:r 8:rr 9: 10: 11: 12:WWWWWW 13:rrr 14: 15: 16: 17:WWW 18: 19:WWWWW 20: 21: 22: 23: 24:rr, W on bar}",
				//     8,13,7,24:
				"{r after playing 6666; !dbl; 1:r 2:r 3: 4: 5: 6:rrrrr 7: 8:rr 9: 10: 11: 12:WWWWWW 13:rrrr 14: 15: 16: 17:WWW 18:r 19:WWWWW 20: 21: 22: 23: 24:r, W on bar}",
				//     8,13,13,13:
				"{r after playing 6666; !dbl; 1:W 2:r 3: 4: 5: 6:rrrrr 7:rrr 8:rr 9: 10: 11: 12:WWWWWW 13:rr 14: 15: 16: 17:WWW 18: 19:WWWWW 20: 21: 22: 23: 24:rr}",
				//     8,13,13,24:
				"{r after playing 6666; !dbl; 1:W 2:r 3: 4: 5: 6:rrrrr 7:rr 8:rr 9: 10: 11: 12:WWWWWW 13:rrr 14: 15: 16: 17:WWW 18:r 19:WWWWW 20: 21: 22: 23: 24:r}",
				//     8,13,24,24:
				"{r after playing 6666; !dbl; 1:W 2:r 3: 4: 5: 6:rrrrr 7:r 8:rr 9: 10: 11: 12:WWWWWW 13:rrrr 14: 15: 16: 17:WWW 18:rr 19:WWWWW 20: 21: 22: 23: 24:}",
				// 13,*,*,*:
				//     13,7,13,7:
				"{r after playing 6666; !dbl; 1:rr 2: 3: 4: 5: 6:rrrrr 7: 8:rrr 9: 10: 11: 12:WWWWWW 13:rrr 14: 15: 16: 17:WWW 18: 19:WWWWW 20: 21: 22: 23: 24:rr, W on bar}",
				//     13,7,13,13:
				"{r after playing 6666; !dbl; 1:r 2: 3: 4: 5: 6:rrrrr 7:rr 8:rrr 9: 10: 11: 12:WWWWWW 13:rr 14: 15: 16: 17:WWW 18: 19:WWWWW 20: 21: 22: 23: 24:rr, W on bar}",
				//     13,7,13,24:
				"{r after playing 6666; !dbl; 1:r 2: 3: 4: 5: 6:rrrrr 7:r 8:rrr 9: 10: 11: 12:WWWWWW 13:rrr 14: 15: 16: 17:WWW 18:r 19:WWWWW 20: 21: 22: 23: 24:r, W on bar}",
				//     13,7,24,24:
				"{r after playing 6666; !dbl; 1:r 2: 3: 4: 5: 6:rrrrr 7: 8:rrr 9: 10: 11: 12:WWWWWW 13:rrrr 14: 15: 16: 17:WWW 18:rr 19:WWWWW 20: 21: 22: 23: 24:, W on bar}",
				//     13,13,13,13:
				"{r after playing 6666; !dbl; 1:W 2: 3: 4: 5: 6:rrrrr 7:rrrr 8:rrr 9: 10: 11: 12:WWWWWW 13:r 14: 15: 16: 17:WWW 18: 19:WWWWW 20: 21: 22: 23: 24:rr}",
				//     13,13,13,24:
				"{r after playing 6666; !dbl; 1:W 2: 3: 4: 5: 6:rrrrr 7:rrr 8:rrr 9: 10: 11: 12:WWWWWW 13:rr 14: 15: 16: 17:WWW 18:r 19:WWWWW 20: 21: 22: 23: 24:r}",
				//     13,13,24,24:
				"{r after playing 6666; !dbl; 1:W 2: 3: 4: 5: 6:rrrrr 7:rr 8:rrr 9: 10: 11: 12:WWWWWW 13:rrr 14: 15: 16: 17:WWW 18:rr 19:WWWWW 20: 21: 22: 23: 24:}",
			}},

		// This example is the mirror image of the one above. The start results
		// from a 65. The roll is 6666.
		example{
			func(board *Board) {
				board.Roller = White
				board.Roll = Roll{6, 6, 6, 6}
				board.Pips[24].Subtract()
				board.Pips[13].Add(Red)
			},
			"{W to play 6666; !dbl; 1:WW 2: 3: 4: 5: 6:rrrrr 7: 8:rrr 9: 10: 11: 12:WWWWW 13:rrrrrr 14: 15: 16: 17:WWW 18: 19:WWWWW 20: 21: 22: 23: 24:r}", // 24 to 18, 18 to 13
			[]string{
				"{W after playing 6666; !dbl; 1: 2: 3: 4: 5: 6:rrrrr 7:WW 8:rrr 9: 10: 11: 12:WWW 13:rrrrrr 14: 15: 16: 17:WWW 18:WW 19:WWWWW 20: 21: 22: 23: 24:r}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4: 5: 6:rrrrr 7:WW 8:rrr 9: 10: 11: 12:WWWW 13:rrrrrr 14: 15: 16: 17:WW 18:W 19:WWWWW 20: 21: 22: 23:W 24:r}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4: 5: 6:rrrrr 7:WW 8:rrr 9: 10: 11: 12:WWWW 13:rrrrrr 14: 15: 16: 17:WWW 18: 19:WWWWW 20: 21: 22: 23: 24:W, r on bar}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4: 5: 6:rrrrr 7:WW 8:rrr 9: 10: 11: 12:WWWWW 13:rrrrrr 14: 15: 16: 17:W 18: 19:WWWWW 20: 21: 22: 23:WW 24:r}",
				"{W after playing 6666; !dbl; 1:W 2: 3: 4: 5: 6:rrrrr 7:W 8:rrr 9: 10: 11: 12:WW 13:rrrrrr 14: 15: 16: 17:WWW 18:WWW 19:WWWWW 20: 21: 22: 23: 24:r}",
				"{W after playing 6666; !dbl; 1:W 2: 3: 4: 5: 6:rrrrr 7:W 8:rrr 9: 10: 11: 12:WWW 13:rrrrrr 14: 15: 16: 17:WW 18:WW 19:WWWWW 20: 21: 22: 23:W 24:r}",
				"{W after playing 6666; !dbl; 1:W 2: 3: 4: 5: 6:rrrrr 7:W 8:rrr 9: 10: 11: 12:WWW 13:rrrrrr 14: 15: 16: 17:WWW 18:W 19:WWWWW 20: 21: 22: 23: 24:W, r on bar}",
				"{W after playing 6666; !dbl; 1:W 2: 3: 4: 5: 6:rrrrr 7:W 8:rrr 9: 10: 11: 12:WWWW 13:rrrrrr 14: 15: 16: 17:W 18:W 19:WWWWW 20: 21: 22: 23:WW 24:r}",
				"{W after playing 6666; !dbl; 1:W 2: 3: 4: 5: 6:rrrrr 7:W 8:rrr 9: 10: 11: 12:WWWW 13:rrrrrr 14: 15: 16: 17:WW 18: 19:WWWWW 20: 21: 22: 23:W 24:W, r on bar}",
				"{W after playing 6666; !dbl; 1:W 2: 3: 4: 5: 6:rrrrr 7:W 8:rrr 9: 10: 11: 12:WWWWW 13:rrrrrr 14: 15: 16: 17: 18: 19:WWWWW 20: 21: 22: 23:WWW 24:r}",
				"{W after playing 6666; !dbl; 1:WW 2: 3: 4: 5: 6:rrrrr 7: 8:rrr 9: 10: 11: 12:W 13:rrrrrr 14: 15: 16: 17:WWW 18:WWWW 19:WWWWW 20: 21: 22: 23: 24:r}",
				"{W after playing 6666; !dbl; 1:WW 2: 3: 4: 5: 6:rrrrr 7: 8:rrr 9: 10: 11: 12:WW 13:rrrrrr 14: 15: 16: 17:WW 18:WWW 19:WWWWW 20: 21: 22: 23:W 24:r}",
				"{W after playing 6666; !dbl; 1:WW 2: 3: 4: 5: 6:rrrrr 7: 8:rrr 9: 10: 11: 12:WW 13:rrrrrr 14: 15: 16: 17:WWW 18:WW 19:WWWWW 20: 21: 22: 23: 24:W, r on bar}",
				"{W after playing 6666; !dbl; 1:WW 2: 3: 4: 5: 6:rrrrr 7: 8:rrr 9: 10: 11: 12:WWW 13:rrrrrr 14: 15: 16: 17:W 18:WW 19:WWWWW 20: 21: 22: 23:WW 24:r}",
				"{W after playing 6666; !dbl; 1:WW 2: 3: 4: 5: 6:rrrrr 7: 8:rrr 9: 10: 11: 12:WWW 13:rrrrrr 14: 15: 16: 17:WW 18:W 19:WWWWW 20: 21: 22: 23:W 24:W, r on bar}",
				"{W after playing 6666; !dbl; 1:WW 2: 3: 4: 5: 6:rrrrr 7: 8:rrr 9: 10: 11: 12:WWW 13:rrrrrr 14: 15: 16: 17:WWW 18: 19:WWWWW 20: 21: 22: 23: 24:WW, r on bar}",
				"{W after playing 6666; !dbl; 1:WW 2: 3: 4: 5: 6:rrrrr 7: 8:rrr 9: 10: 11: 12:WWWW 13:rrrrrr 14: 15: 16: 17: 18:W 19:WWWWW 20: 21: 22: 23:WWW 24:r}",
				"{W after playing 6666; !dbl; 1:WW 2: 3: 4: 5: 6:rrrrr 7: 8:rrr 9: 10: 11: 12:WWWW 13:rrrrrr 14: 15: 16: 17:W 18: 19:WWWWW 20: 21: 22: 23:WW 24:W, r on bar}",
			}},

		// Tests a wasted 6666 where a naive implementation would bear off 4.
		example{
			func(board *Board) {
				board.Roller = White
				board.Roll = Roll{6, 6, 6, 6}
				board.Pips = Points28{}
				board.Pips[18].Reset(4, White)
				board.Pips[19].Reset(4, White)
				board.Pips[20].Reset(4, White)
				board.Pips[21].Reset(3, White)
				board.Pips[24].Reset(2, Red)
				board.Pips[BorneOffRedPip].Reset(13, Red)
				// {W to play 6666; !dbl; 1:WW 2: 3: 4: 5: 6:rrrrr 7: 8:rrr 9: 10: 11: 12:WWWWW 13:rrrrr 14: 15: 16: 17:WWW 18: 19:WWWWW 20: 21: 22: 23: 24:rr}
			},
			"{W to play 6666; !dbl; 1: 2: 3: 4: 5: 6: 7: 8: 9: 10: 11: 12: 13: 14: 15: 16: 17: 18:WWWW 19:WWWW 20:WWWW 21:WWW 22: 23: 24:rr, 13 r off}",
			[]string{"{W to play 6666; !dbl; 1: 2: 3: 4: 5: 6: 7: 8: 9: 10: 11: 12: 13: 14: 15: 16: 17: 18:WWWW 19:WWWW 20:WWWW 21:WWW 22: 23: 24:rr, 13 r off}"}},

		// Mirror image of the above to test both Red and White code paths.
		example{
			func(board *Board) {
				board.Roller = Red
				board.Roll = Roll{6, 6, 6, 6}
				board.Pips = Points28{}
				board.Pips[7].Reset(4, Red)
				board.Pips[6].Reset(4, Red)
				board.Pips[5].Reset(4, Red)
				board.Pips[4].Reset(3, Red)
				board.Pips[1].Reset(2, White)
				board.Pips[BorneOffWhitePip].Reset(13, White)
				// {W to play 6666; !dbl; 1:WW 2: 3: 4: 5: 6:rrrrr 7: 8:rrr 9: 10: 11: 12:WWWWW 13:rrrrr 14: 15: 16: 17:WWW 18: 19:WWWWW 20: 21: 22: 23: 24:rr}
			},
			"{r to play 6666; !dbl; 1:WW 2: 3: 4:rrr 5:rrrr 6:rrrr 7:rrrr 8: 9: 10: 11: 12: 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, 13 W off}",
			[]string{"{r to play 6666; !dbl; 1:WW 2: 3: 4:rrr 5:rrrr 6:rrrr 7:rrrr 8: 9: 10: 11: 12: 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, 13 W off}"}},

		example{
			func(board *Board) {
				board.Roller = Red
				board.Roll = Roll{6, 6, 6, 6}
				board.Pips = Points28{}
				board.Pips[7].Reset(3, Red)
				board.Pips[6].Reset(4, Red)
				board.Pips[5].Reset(4, Red)
				board.Pips[4].Reset(4, Red)
				board.Pips[1].Reset(1, White)
				board.Pips[BorneOffWhitePip].Reset(14, White)
			},
			"{r to play 6666; !dbl; 1:W 2: 3: 4:rrrr 5:rrrr 6:rrrr 7:rrr 8: 9: 10: 11: 12: 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, 14 W off}",
			[]string{"{r after playing 6666; !dbl; 1:rrr 2: 3: 4:rrrr 5:rrrr 6:rrr 7: 8: 9: 10: 11: 12: 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, W on bar, 14 W off, 1 r off}"}},

		// mirror of the above
		example{
			func(board *Board) {
				board.Roller = White
				board.Roll = Roll{6, 6, 6, 6}
				board.Pips = Points28{}
				board.Pips[18].Reset(3, White)
				board.Pips[19].Reset(4, White)
				board.Pips[20].Reset(4, White)
				board.Pips[21].Reset(4, White)
				board.Pips[24].Reset(1, Red)
				board.Pips[BorneOffRedPip].Reset(14, Red)
			},
			"{W to play 6666; !dbl; 1: 2: 3: 4: 5: 6: 7: 8: 9: 10: 11: 12: 13: 14: 15: 16: 17: 18:WWW 19:WWWW 20:WWWW 21:WWWW 22: 23: 24:r, 14 r off}",
			[]string{"{W after playing 6666; !dbl; 1: 2: 3: 4: 5: 6: 7: 8: 9: 10: 11: 12: 13: 14: 15: 16: 17: 18: 19:WWW 20:WWWW 21:WWWW 22: 23: 24:WWW, r on bar, 1 W off, 14 r off}"}},

		// 96 candidates, 96 times the fun!
		example{
			func(board *Board) {
				board.Roller = White
				board.Roll = Roll{6, 6, 6, 6}
				board.Pips = Points28{}
				board.Pips[7].Reset(3, White)
				board.Pips[6].Reset(4, White)
				board.Pips[5].Reset(4, White)
				board.Pips[4].Reset(4, White)
				board.Pips[BarRedPip].Reset(1, Red)
				board.Pips[BorneOffRedPip].Reset(14, Red)
			},
			"{W to play 6666; !dbl; 1: 2: 3: 4:WWWW 5:WWWW 6:WWWW 7:WWW 8: 9: 10: 11: 12: 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
			// the following appear once and only once: 17:WW 18:WW 19:WW 6:"" 5:"" 4:""
			[]string{
				"{W after playing 6666; !dbl; 1: 2: 3: 4: 5:WWWW 6:WWWW 7:WWW 8: 9: 10:WWWW 11: 12: 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:W 5:WWW 6:WWWW 7:WWW 8: 9: 10:WWW 11:W 12: 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:W 5:WWWW 6:WWW 7:WWW 8: 9: 10:WWW 11: 12:W 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:W 5:WWWW 6:WWWW 7:WW 8: 9: 10:WWW 11: 12: 13:W 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:W 5:WWWW 6:WWWW 7:WWW 8: 9: 10:WW 11: 12: 13: 14: 15: 16:W 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WW 5:WW 6:WWWW 7:WWW 8: 9: 10:WW 11:WW 12: 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WW 5:WWW 6:WWW 7:WWW 8: 9: 10:WW 11:W 12:W 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WW 5:WWW 6:WWWW 7:WW 8: 9: 10:WW 11:W 12: 13:W 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WW 5:WWW 6:WWWW 7:WWW 8: 9: 10:W 11:W 12: 13: 14: 15: 16:W 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WW 5:WWW 6:WWWW 7:WWW 8: 9: 10:WW 11: 12: 13: 14: 15: 16: 17:W 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WW 5:WWWW 6:WW 7:WWW 8: 9: 10:WW 11: 12:WW 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WW 5:WWWW 6:WWW 7:WW 8: 9: 10:WW 11: 12:W 13:W 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WW 5:WWWW 6:WWW 7:WWW 8: 9: 10:W 11: 12:W 13: 14: 15: 16:W 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WW 5:WWWW 6:WWW 7:WWW 8: 9: 10:WW 11: 12: 13: 14: 15: 16: 17: 18:W 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WW 5:WWWW 6:WWWW 7:W 8: 9: 10:WW 11: 12: 13:WW 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WW 5:WWWW 6:WWWW 7:WW 8: 9: 10:W 11: 12: 13:W 14: 15: 16:W 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WW 5:WWWW 6:WWWW 7:WW 8: 9: 10:WW 11: 12: 13: 14: 15: 16: 17: 18: 19:W 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WW 5:WWWW 6:WWWW 7:WWW 8: 9: 10: 11: 12: 13: 14: 15: 16:WW 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WW 5:WWWW 6:WWWW 7:WWW 8: 9: 10:W 11: 12: 13: 14: 15: 16: 17: 18: 19: 20: 21: 22:W 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWW 5:W 6:WWWW 7:WWW 8: 9: 10:W 11:WWW 12: 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWW 5:WW 6:WWW 7:WWW 8: 9: 10:W 11:WW 12:W 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWW 5:WW 6:WWWW 7:WW 8: 9: 10:W 11:WW 12: 13:W 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWW 5:WW 6:WWWW 7:WWW 8: 9: 10: 11:WW 12: 13: 14: 15: 16:W 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWW 5:WW 6:WWWW 7:WWW 8: 9: 10:W 11:W 12: 13: 14: 15: 16: 17:W 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWW 5:WWW 6:WW 7:WWW 8: 9: 10:W 11:W 12:WW 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWW 5:WWW 6:WWW 7:WW 8: 9: 10:W 11:W 12:W 13:W 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWW 5:WWW 6:WWW 7:WWW 8: 9: 10: 11:W 12:W 13: 14: 15: 16:W 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWW 5:WWW 6:WWW 7:WWW 8: 9: 10:W 11: 12:W 13: 14: 15: 16: 17:W 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWW 5:WWW 6:WWW 7:WWW 8: 9: 10:W 11:W 12: 13: 14: 15: 16: 17: 18:W 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWW 5:WWW 6:WWWW 7:W 8: 9: 10:W 11:W 12: 13:WW 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWW 5:WWW 6:WWWW 7:WW 8: 9: 10: 11:W 12: 13:W 14: 15: 16:W 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWW 5:WWW 6:WWWW 7:WW 8: 9: 10:W 11: 12: 13:W 14: 15: 16: 17:W 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWW 5:WWW 6:WWWW 7:WW 8: 9: 10:W 11:W 12: 13: 14: 15: 16: 17: 18: 19:W 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWW 5:WWW 6:WWWW 7:WWW 8: 9: 10: 11: 12: 13: 14: 15: 16:W 17:W 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWW 5:WWW 6:WWWW 7:WWW 8: 9: 10: 11:W 12: 13: 14: 15: 16: 17: 18: 19: 20: 21: 22:W 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWW 5:WWW 6:WWWW 7:WWW 8: 9: 10:W 11: 12: 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23:W 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWW 5:WWWW 6:W 7:WWW 8: 9: 10:W 11: 12:WWW 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWW 5:WWWW 6:WW 7:WW 8: 9: 10:W 11: 12:WW 13:W 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWW 5:WWWW 6:WW 7:WWW 8: 9: 10: 11: 12:WW 13: 14: 15: 16:W 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWW 5:WWWW 6:WW 7:WWW 8: 9: 10:W 11: 12:W 13: 14: 15: 16: 17: 18:W 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWW 5:WWWW 6:WWW 7:W 8: 9: 10:W 11: 12:W 13:WW 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWW 5:WWWW 6:WWW 7:WW 8: 9: 10: 11: 12:W 13:W 14: 15: 16:W 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWW 5:WWWW 6:WWW 7:WW 8: 9: 10:W 11: 12: 13:W 14: 15: 16: 17: 18:W 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWW 5:WWWW 6:WWW 7:WW 8: 9: 10:W 11: 12:W 13: 14: 15: 16: 17: 18: 19:W 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWW 5:WWWW 6:WWW 7:WWW 8: 9: 10: 11: 12: 13: 14: 15: 16:W 17: 18:W 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWW 5:WWWW 6:WWW 7:WWW 8: 9: 10: 11: 12:W 13: 14: 15: 16: 17: 18: 19: 20: 21: 22:W 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWW 5:WWWW 6:WWW 7:WWW 8: 9: 10:W 11: 12: 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:W, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWW 5:WWWW 6:WWWW 7: 8: 9: 10:W 11: 12: 13:WWW 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWW 5:WWWW 6:WWWW 7:W 8: 9: 10: 11: 12: 13:WW 14: 15: 16:W 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWW 5:WWWW 6:WWWW 7:W 8: 9: 10:W 11: 12: 13:W 14: 15: 16: 17: 18: 19:W 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWW 5:WWWW 6:WWWW 7:WW 8: 9: 10: 11: 12: 13: 14: 15: 16:W 17: 18: 19:W 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWW 5:WWWW 6:WWWW 7:WW 8: 9: 10: 11: 12: 13:W 14: 15: 16: 17: 18: 19: 20: 21: 22:W 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5: 6:WWWW 7:WWW 8: 9: 10: 11:WWWW 12: 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:W 6:WWW 7:WWW 8: 9: 10: 11:WWW 12:W 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:W 6:WWWW 7:WW 8: 9: 10: 11:WWW 12: 13:W 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:W 6:WWWW 7:WWW 8: 9: 10: 11:WW 12: 13: 14: 15: 16: 17:W 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:WW 6:WW 7:WWW 8: 9: 10: 11:WW 12:WW 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:WW 6:WWW 7:WW 8: 9: 10: 11:WW 12:W 13:W 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:WW 6:WWW 7:WWW 8: 9: 10: 11:W 12:W 13: 14: 15: 16: 17:W 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:WW 6:WWW 7:WWW 8: 9: 10: 11:WW 12: 13: 14: 15: 16: 17: 18:W 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:WW 6:WWWW 7:W 8: 9: 10: 11:WW 12: 13:WW 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:WW 6:WWWW 7:WW 8: 9: 10: 11:W 12: 13:W 14: 15: 16: 17:W 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:WW 6:WWWW 7:WW 8: 9: 10: 11:WW 12: 13: 14: 15: 16: 17: 18: 19:W 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:WW 6:WWWW 7:WWW 8: 9: 10: 11: 12: 13: 14: 15: 16: 17:WW 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:WW 6:WWWW 7:WWW 8: 9: 10: 11:W 12: 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23:W 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:WWW 6:W 7:WWW 8: 9: 10: 11:W 12:WWW 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:WWW 6:WW 7:WW 8: 9: 10: 11:W 12:WW 13:W 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:WWW 6:WW 7:WWW 8: 9: 10: 11: 12:WW 13: 14: 15: 16: 17:W 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:WWW 6:WW 7:WWW 8: 9: 10: 11:W 12:W 13: 14: 15: 16: 17: 18:W 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:WWW 6:WWW 7:W 8: 9: 10: 11:W 12:W 13:WW 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:WWW 6:WWW 7:WW 8: 9: 10: 11: 12:W 13:W 14: 15: 16: 17:W 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:WWW 6:WWW 7:WW 8: 9: 10: 11:W 12: 13:W 14: 15: 16: 17: 18:W 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:WWW 6:WWW 7:WW 8: 9: 10: 11:W 12:W 13: 14: 15: 16: 17: 18: 19:W 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:WWW 6:WWW 7:WWW 8: 9: 10: 11: 12: 13: 14: 15: 16: 17:W 18:W 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:WWW 6:WWW 7:WWW 8: 9: 10: 11: 12:W 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23:W 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:WWW 6:WWW 7:WWW 8: 9: 10: 11:W 12: 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:W, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:WWW 6:WWWW 7: 8: 9: 10: 11:W 12: 13:WWW 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:WWW 6:WWWW 7:W 8: 9: 10: 11: 12: 13:WW 14: 15: 16: 17:W 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:WWW 6:WWWW 7:W 8: 9: 10: 11:W 12: 13:W 14: 15: 16: 17: 18: 19:W 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:WWW 6:WWWW 7:WW 8: 9: 10: 11: 12: 13: 14: 15: 16: 17:W 18: 19:W 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:WWW 6:WWWW 7:WW 8: 9: 10: 11: 12: 13:W 14: 15: 16: 17: 18: 19: 20: 21: 22: 23:W 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:WWWW 6: 7:WWW 8: 9: 10: 11: 12:WWWW 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:WWWW 6:W 7:WW 8: 9: 10: 11: 12:WWW 13:W 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:WWWW 6:W 7:WWW 8: 9: 10: 11: 12:WW 13: 14: 15: 16: 17: 18:W 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:WWWW 6:WW 7:W 8: 9: 10: 11: 12:WW 13:WW 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:WWWW 6:WW 7:WW 8: 9: 10: 11: 12:W 13:W 14: 15: 16: 17: 18:W 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:WWWW 6:WW 7:WW 8: 9: 10: 11: 12:WW 13: 14: 15: 16: 17: 18: 19:W 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:WWWW 6:WW 7:WWW 8: 9: 10: 11: 12: 13: 14: 15: 16: 17: 18:WW 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:WWWW 6:WW 7:WWW 8: 9: 10: 11: 12:W 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:W, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:WWWW 6:WWW 7: 8: 9: 10: 11: 12:W 13:WWW 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:WWWW 6:WWW 7:W 8: 9: 10: 11: 12: 13:WW 14: 15: 16: 17: 18:W 19: 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:WWWW 6:WWW 7:W 8: 9: 10: 11: 12:W 13:W 14: 15: 16: 17: 18: 19:W 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:WWWW 6:WWW 7:WW 8: 9: 10: 11: 12: 13: 14: 15: 16: 17: 18:W 19:W 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:WWWW 6:WWW 7:WW 8: 9: 10: 11: 12: 13:W 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:W, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:WWWW 6:WWWW 7: 8: 9: 10: 11: 12: 13:WW 14: 15: 16: 17: 18: 19:W 20: 21: 22: 23: 24:, r on bar, 14 r off}",
				"{W after playing 6666; !dbl; 1: 2: 3: 4:WWWW 5:WWWW 6:WWWW 7:W 8: 9: 10: 11: 12: 13: 14: 15: 16: 17: 18: 19:WW 20: 21: 22: 23: 24:, r on bar, 14 r off}",
			}},

		example{
			func(board *Board) {
				board.Roller = Red
				board.Roll = Roll{6, 6, 6, 6}
				board.Pips = Points28{}
				board.Pips[7].Reset(3, Red)
				board.Pips[5].Reset(8, Red)
				board.Pips[4].Reset(4, Red)
				board.Pips[1].Reset(1, White)
				board.Pips[BorneOffWhitePip].Reset(14, White)
			},
			"{r to play 6666; !dbl; 1:W 2: 3: 4:rrrr 5:rrrrrrrr 6: 7:rrr 8: 9: 10: 11: 12: 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, 14 W off}",
			[]string{"{r after playing 6666; !dbl; 1:rrr 2: 3: 4:rrrr 5:rrrrrrr 6: 7: 8: 9: 10: 11: 12: 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, W on bar, 14 W off, 1 r off}"}},

		// mirror of the above
		example{
			func(board *Board) {
				board.Roller = White
				board.Roll = Roll{6, 6, 6, 6}
				board.Pips = Points28{}
				board.Pips[18].Reset(3, White)
				board.Pips[20].Reset(8, White)
				board.Pips[21].Reset(4, White)
				board.Pips[24].Reset(1, Red)
				board.Pips[BorneOffRedPip].Reset(14, Red)
			},
			"{W to play 6666; !dbl; 1: 2: 3: 4: 5: 6: 7: 8: 9: 10: 11: 12: 13: 14: 15: 16: 17: 18:WWW 19: 20:WWWWWWWW 21:WWWW 22: 23: 24:r, 14 r off}",
			[]string{"{W after playing 6666; !dbl; 1: 2: 3: 4: 5: 6: 7: 8: 9: 10: 11: 12: 13: 14: 15: 16: 17: 18: 19: 20:WWWWWWW 21:WWWW 22: 23: 24:WWW, r on bar, 1 W off, 14 r off}"}},

		example{
			func(board *Board) {
				board.Roller = Red
				board.Roll = Roll{6, 6, 6, 6}
				board.Pips = Points28{}
				board.Pips[7].Reset(2, Red)
				board.Pips[8].Reset(1, White)
				board.Pips[BorneOffWhitePip].Reset(14, White)
				board.Pips[BorneOffRedPip].Reset(13, Red)
			},
			"{r to play 6666; !dbl; 1: 2: 3: 4: 5: 6: 7:rr 8:W 9: 10: 11: 12: 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, 14 W off, 13 r off}",
			[]string{"{r after playing 6666; !dbl; 1: 2: 3: 4: 5: 6: 7: 8:W 9: 10: 11: 12: 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, 14 W off, 15 r off}"}},

		// mirror of the above
		example{
			func(board *Board) {
				board.Roller = White
				board.Roll = Roll{6, 6, 6, 6}
				board.Pips = Points28{}
				board.Pips[17].Reset(2, Red)
				board.Pips[16].Reset(1, White)
				board.Pips[BorneOffWhitePip].Reset(14, White)
				board.Pips[BorneOffRedPip].Reset(13, Red)
			},
			"{W to play 6666; !dbl; 1: 2: 3: 4: 5: 6: 7: 8: 9: 10: 11: 12: 13: 14: 15: 16:W 17:rr 18: 19: 20: 21: 22: 23: 24:, 14 W off, 13 r off}",
			[]string{"{W to play   66 after playing   66; !dbl; 1: 2: 3: 4: 5: 6: 7: 8: 9: 10: 11: 12: 13: 14: 15: 16: 17:rr 18: 19: 20: 21: 22: 23: 24:, 15 W off, 13 r off}"}},

		example{
			func(board *Board) {
				board.Roller = Red
				board.Roll = Roll{6, 6, 6, 6}
				board.Pips = Points28{}
				board.Pips[7].Reset(1, Red)
				board.Pips[8].Reset(1, White)
				board.Pips[BorneOffWhitePip].Reset(14, White)
				board.Pips[BorneOffRedPip].Reset(14, Red)
			},
			"{r to play 6666; !dbl; 1: 2: 3: 4: 5: 6: 7:r 8:W 9: 10: 11: 12: 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, 14 W off, 14 r off}",
			[]string{"{r to play   66 after playing   66; !dbl; 1: 2: 3: 4: 5: 6: 7: 8:W 9: 10: 11: 12: 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, 14 W off, 15 r off}"}},

		// Mirror of the above
		example{
			func(board *Board) {
				board.Roller = White
				board.Roll = Roll{6, 6, 6, 6}
				board.Pips = Points28{}
				board.Pips[17].Reset(1, White)
				board.Pips[16].Reset(1, Red)
				board.Pips[BorneOffWhitePip].Reset(14, White)
				board.Pips[BorneOffRedPip].Reset(14, Red)
			},
			"{W to play 6666; !dbl; 1: 2: 3: 4: 5: 6: 7: 8: 9: 10: 11: 12: 13: 14: 15: 16:r 17:W 18: 19: 20: 21: 22: 23: 24:, 14 W off, 14 r off}",
			[]string{"{W to play   66 after playing   66; !dbl; 1: 2: 3: 4: 5: 6: 7: 8: 9: 10: 11: 12: 13: 14: 15: 16:r 17: 18: 19: 20: 21: 22: 23: 24:, 15 W off, 14 r off}"}},

		// A roll of 63 where you cannot use both but you can use either one
		// and thus are forced to use the 6 because of a backgammon rule "you
		// must use the larger one if you can use either but not both"
		example{
			func(board *Board) {
				board.Roller = White
				board.Roll = Roll{6, 3}
				board.Pips = Points28{}
				board.Pips[9].Reset(2, Red)
				board.Pips[BarWhitePip].Reset(1, White)
				board.Pips[BorneOffWhitePip].Reset(14, White)
				board.Pips[BorneOffRedPip].Reset(13, Red)
			},
			"{W to play   63; !dbl; 1: 2: 3: 4: 5: 6: 7: 8: 9:rr 10: 11: 12: 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, W on bar, 14 W off, 13 r off}",
			[]string{"{W to play    3 after playing    6; !dbl; 1: 2: 3: 4: 5: 6:W 7: 8: 9:rr 10: 11: 12: 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, 14 W off, 13 r off}"}},

		// A roll of 63 where you could quasilegally use just the 3 but have
		// another play that uses both.
		example{
			func(board *Board) {
				board.Roller = Red
				board.Roll = Roll{6, 3}
				board.Pips = Points28{}
				board.Pips[11].Reset(1, Red)
				board.Pips[6].Reset(2, Red)
				board.Pips[2].Reset(2, White)
				board.Pips[BarWhitePip].Reset(13, White)
				board.Pips[BorneOffRedPip].Reset(12, Red)
			},
			"{r to play   63; !dbl; 1: 2:WW 3: 4: 5: 6:rr 7: 8: 9: 10: 11:r 12: 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, WWWWWWWWWWWWW on bar, 12 r off}",
			[]string{
				"{r after playing   63; !dbl; 1: 2:WW 3:r 4: 5:r 6:r 7: 8: 9: 10: 11: 12: 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, WWWWWWWWWWWWW on bar, 12 r off}"}},
	}
	for exNum, ex := range examples {
		board := New(true)
		ex.Initializer(board)
		if i := board.Invalidity(EnforceRollValidity); i != "" {
			t.Errorf("exNum=%d: invalidity is %v", exNum, i)
			continue
		}
		if q := board.String(); q != ex.InitializerCheck {
			t.Errorf(
				"exNum=%d; Initializercheck failed q=\n%v\nwhen we expected\n%v",
				exNum, q, ex.InitializerCheck)
			continue
		}
		candidates := board.LegalContinuations()
		if len(candidates) != len(ex.continuations) {
			t.Errorf(
				"exNum=%d (%v):\nexpected %d candidates, not %d:\n%v",
				exNum, ex.InitializerCheck, len(ex.continuations), len(candidates),
				prettyCandidates(candidates))
		} else {
			for i, v := range ex.continuations {
				if iv := candidates[i].Invalidity(IgnoreRollValidity); iv != "" {
					t.Errorf("invalidity=%v", iv)
				}
				if s := candidates[i].String(); s != v {
					t.Errorf(
						"exNum=%d\nstart:%v\ncandidates[%d] is\n%v\nnot\n%v\nas expected",
						exNum, ex.InitializerCheck, i, s, v)
				}
			}
		}
	}
}

func assertValidity(board *Board, t *testing.T) {
	if i := board.Invalidity(EnforceRollValidity); i != "" {
		panic(fmt.Sprintf("invalidity: %v in board %v", i, board))
	}
}

func TestComingOffTheBar(t *testing.T) {
	type example struct {
		Initializer      func(*Board) // is passed a board from New()
		InitializerCheck string
		continuations    []string
	}
	freshBoard := "{W to play   65; !dbl; 1:WW 2: 3: 4: 5: 6:rrrrr 7: 8:rrr 9: 10: 11: 12:WWWWW 13:rrrrr 14: 15: 16: 17:WWW 18: 19:WWWWW 20: 21: 22: 23: 24:rr}"
	examples := [...]example{
		example{
			func(board *Board) {
				board.Roller = White
				board.Roll = Roll{6, 5} // 5 is empty but 6 is blocked by 5 Red
				board.Pips[1].Reset(1, White)
				board.Pips[BarWhitePip].Add(White)
			},
			"{W to play   65; !dbl; 1:W 2: 3: 4: 5: 6:rrrrr 7: 8:rrr 9: 10: 11: 12:WWWWW 13:rrrrr 14: 15: 16: 17:WWW 18: 19:WWWWW 20: 21: 22: 23: 24:rr, W on bar}",
			[]string{
				"{W to play    6 after playing    5; !dbl; 1:W 2: 3: 4: 5:W 6:rrrrr 7: 8:rrr 9: 10: 11: 12:WWWWW 13:rrrrr 14: 15: 16: 17:WWW 18: 19:WWWWW 20: 21: 22: 23: 24:rr}"}},

		example{
			func(board *Board) {
				board.Roller = White
				board.Roll = Roll{6, 5} // 5 is empty but 6 is blocked by 5 Red
			},
			freshBoard,
			nil},

		example{
			func(board *Board) {
				board.Roller = White
				board.Roll = Roll{6, 6, 6, 6}
				board.Pips[1].Reset(1, White)
				board.Pips[BarWhitePip].Add(White)
			},
			"{W to play 6666; !dbl; 1:W 2: 3: 4: 5: 6:rrrrr 7: 8:rrr 9: 10: 11: 12:WWWWW 13:rrrrr 14: 15: 16: 17:WWW 18: 19:WWWWW 20: 21: 22: 23: 24:rr, W on bar}",
			nil},

		example{
			func(board *Board) {
				board.Roller = White
				board.Roll = Roll{1, 1, 1, 1}
				board.Pips[19].Reset(0, White)
				board.Pips[BarWhitePip].Reset(5, White)
			},
			"{W to play 1111; !dbl; 1:WW 2: 3: 4: 5: 6:rrrrr 7: 8:rrr 9: 10: 11: 12:WWWWW 13:rrrrr 14: 15: 16: 17:WWW 18: 19: 20: 21: 22: 23: 24:rr, WWWWW on bar}",
			[]string{"{W after playing 1111; !dbl; 1:WWWWWW 2: 3: 4: 5: 6:rrrrr 7: 8:rrr 9: 10: 11: 12:WWWWW 13:rrrrr 14: 15: 16: 17:WWW 18: 19: 20: 21: 22: 23: 24:rr, W on bar}"}},

		example{
			func(board *Board) {
				board.Roller = White
				board.Roll = Roll{1, 1, 1, 1}
				board.Pips[19].Reset(3, White)
				board.Pips[BarWhitePip].Reset(2, White)
			},
			"{W to play 1111; !dbl; 1:WW 2: 3: 4: 5: 6:rrrrr 7: 8:rrr 9: 10: 11: 12:WWWWW 13:rrrrr 14: 15: 16: 17:WWW 18: 19:WWW 20: 21: 22: 23: 24:rr, WW on bar}",
			[]string{"{W to play   11 after playing   11; !dbl; 1:WWWW 2: 3: 4: 5: 6:rrrrr 7: 8:rrr 9: 10: 11: 12:WWWWW 13:rrrrr 14: 15: 16: 17:WWW 18: 19:WWW 20: 21: 22: 23: 24:rr}"}},

		example{
			func(board *Board) {
				board.Roller = White
				board.Roll = Roll{1, 1, 1, 1}
				board.Pips[19].Reset(2, White)
				board.Pips[BarWhitePip].Reset(3, White)
			},
			"{W to play 1111; !dbl; 1:WW 2: 3: 4: 5: 6:rrrrr 7: 8:rrr 9: 10: 11: 12:WWWWW 13:rrrrr 14: 15: 16: 17:WWW 18: 19:WW 20: 21: 22: 23: 24:rr, WWW on bar}",
			[]string{"{W to play    1 after playing  111; !dbl; 1:WWWWW 2: 3: 4: 5: 6:rrrrr 7: 8:rrr 9: 10: 11: 12:WWWWW 13:rrrrr 14: 15: 16: 17:WWW 18: 19:WW 20: 21: 22: 23: 24:rr}"}},

		example{
			func(board *Board) {
				board.Roller = White
				board.Roll = Roll{6, 6, 6, 6}
				board.Pips[19].Reset(0, White)
				board.Pips[BarWhitePip].Reset(5, White)
				board.Pips[6].Reset(1, Red)
				board.Pips[BarRedPip].Reset(4, Red)
			},
			"{W to play 6666; !dbl; 1:WW 2: 3: 4: 5: 6:r 7: 8:rrr 9: 10: 11: 12:WWWWW 13:rrrrr 14: 15: 16: 17:WWW 18: 19: 20: 21: 22: 23: 24:rr, WWWWW on bar, rrrr on bar}",
			[]string{"{W after playing 6666; !dbl; 1:WW 2: 3: 4: 5: 6:WWWW 7: 8:rrr 9: 10: 11: 12:WWWWW 13:rrrrr 14: 15: 16: 17:WWW 18: 19: 20: 21: 22: 23: 24:rr, W on bar, rrrrr on bar}"}},

		example{
			func(board *Board) {
				board.Roller = White
				board.Roll = Roll{6, 5}
				board.Pips[1].Reset(1, White)
				board.Pips[BarWhitePip].Reset(1, White)
				board.Pips[6].Reset(1, Red)
				board.Pips[BarRedPip].Reset(4, Red)
			},
			"{W to play   65; !dbl; 1:W 2: 3: 4: 5: 6:r 7: 8:rrr 9: 10: 11: 12:WWWWW 13:rrrrr 14: 15: 16: 17:WWW 18: 19:WWWWW 20: 21: 22: 23: 24:rr, W on bar, rrrr on bar}",
			[]string{
				"{W to play    5 after playing    6; !dbl; 1:W 2: 3: 4: 5: 6:W 7: 8:rrr 9: 10: 11: 12:WWWWW 13:rrrrr 14: 15: 16: 17:WWW 18: 19:WWWWW 20: 21: 22: 23: 24:rr, rrrrr on bar}",
				"{W to play    6 after playing    5; !dbl; 1:W 2: 3: 4: 5:W 6:r 7: 8:rrr 9: 10: 11: 12:WWWWW 13:rrrrr 14: 15: 16: 17:WWW 18: 19:WWWWW 20: 21: 22: 23: 24:rr, rrrr on bar}"}},
	}

	for exNum, ex := range examples {
		board := New(true)
		ex.Initializer(board)
		if iv := board.Invalidity(EnforceRollValidity); iv != "" {
			t.Fatalf("exNum=%d ex=%v iv=%v", exNum, ex, iv)
		}
		if y := board.String(); y != ex.InitializerCheck {
			t.Fatalf(
				"bad initializer %d: expected %v but got %v",
				exNum, ex.InitializerCheck, y)
		}
		candidates := board.continuationsOffTheBar()
		if expected := len(ex.continuations); len(candidates) != expected {
			t.Errorf(
				"exNum=%d expected %d candidates, not %v\nexpected: %v",
				exNum, expected, candidates, ex.continuations)
		} else {
			for d, continuation := range ex.continuations {
				if candidates[d].String() != continuation {
					t.Errorf(
						"exNum=%d candidates[%d] is \n%v\nand we wanted\n%v",
						exNum, d, candidates[d], continuation)
				}
			}
		}
	}
}

func TestRollUse(t *testing.T) {
	type example struct {
		Begin Roll
		End   Roll
		Die   Die
	}
	examples := [...]example{
		example{Roll{1}, Roll{}, 1},
		example{Roll{1, 1}, Roll{1}, 1},
		example{Roll{1, 1, 1, 1}, Roll{1, 1, 1}, 1},
	}
	for _, e := range examples {
		tmp := Roll{1, 1, 1}
		if x := e.Begin.Use(e.Die, &tmp); x != e.End {
			t.Errorf("example is %v; actual End is %v; expected End is %v", e, x, e.End)
		}
		expectedTmp := Roll{1, 1, 1, e.Die}
		if tmp != expectedTmp {
			t.Errorf("example is %v; tmp is %v", e, tmp)
		}
	}
}

func dieSliceEquals(a, b []Die) bool {
	if len(a) != len(b) {
		return false
	}
	for i, aa := range a {
		if aa != b[i] {
			return false
		}
	}
	return true
}

func TestRollUniqueDice(t *testing.T) {
	type example struct {
		Roll   Roll
		Result []Die
	}
	examples := [...]example{
		example{Roll{}, []Die{}},
		example{Roll{1}, []Die{1}},
		example{Roll{1, 1, 1, 1}, []Die{1}},
		example{Roll{1, 6}, []Die{6, 1}},
		example{Roll{1, 6, 4}, []Die{6, 4, 1}}, // invalid Roll, though
	}
	for _, e := range examples {
		if x := e.Roll.UniqueDice(); !dieSliceEquals(e.Result, x) {
			t.Errorf("example is %v; result is %v", e, x)
		}
	}
}

func TestRollEquals(t *testing.T) {
	r1 := Roll{1}
	r2 := Roll{2}
	r11 := Roll{1, 1}
	r65 := Roll{6, 5}
	r56 := Roll{5, 6}
	if r1.Equals(r11) || r1.Equals(r2) || r1.Equals(r65) {
		t.Errorf("oops")
	}
	if !r1.Equals(r1) {
		t.Errorf("oops")
	}
	if r11.Equals(r65) || r56.Equals(r11) {
		t.Errorf("oops")
	}
	if !r65.Equals(r65) || !r65.Equals(r56) || !r56.Equals(r56) || !r56.Equals(r65) {
		t.Errorf("oops")
	}
}

func TestBoardMemoryFootprint(t *testing.T) {
	if s := unsafe.Sizeof(*New(true)); s != 88 {
		pair := runtime.GOOS + "-" + runtime.GOARCH
		t.Fatalf(
			"sizeof(Board) on %s is %d. This is not necessarily a problem, but you run the benchmarks again with `make bench`",
			pair, s)
	}
}

func TestTakeTurnSingleStakes(t *testing.T) {
	board := New(true)
	board.Roller = White
	board.Roll = Roll{6, 6, 6, 6}
	board.Pips = Points28{}
	board.Pips[17].Reset(2, Red)
	board.Pips[16].Reset(1, White)
	board.Pips[BorneOffWhitePip].Reset(14, White)
	board.Pips[BorneOffRedPip].Reset(13, Red)
	assertValidity(board, t)
	next := board.LegalContinuations()
	if len(next) != 1 {
		t.Fatalf("TestLegalContinuations has this guy... %v", next)
	}
	if s := next[0].String(); s != "{W to play   66 after playing   66; !dbl; 1: 2: 3: 4: 5: 6: 7: 8: 9: 10: 11: 12: 13: 14: 15: 16: 17:rr 18: 19: 20: 21: 22: 23: 24:, 15 W off, 13 r off}" {
		t.Fatalf("TestLegalContinuations has this guy... %v", s)
	}
	victor, stakes, score := next[0].TakeTurn(nil, nil)
	if victor != White || stakes != 1 || score.String() != "Score{Goal:0,W:1,r:0,Crawford on,inactive}" {
		t.Errorf("victor=%v stakes=%v score=%v", victor, stakes, score)
	}
}

func TestTakeTurnRedAcceptsDouble(t *testing.T) {
	rand.Seed(37)
	board := New(true)
	board.Roller = Red
	board.Roll = Roll{6, 6, 6, 6}
	board.Pips = Points28{}
	board.Pips[17].Reset(10, Red)
	board.Pips[16].Reset(10, White)
	board.Pips[BorneOffWhitePip].Reset(5, White)
	board.Pips[BorneOffRedPip].Reset(5, Red)
	assertValidity(board, t)
	next := board.LegalContinuations()
	victor, stakes, _ := next[0].TakeTurn(
		func(_ *Board) bool { return true },
		func(_ *Board) bool { return true })
	if victor != NoChecker || stakes != 0 {
		t.Errorf("victor=%v stakes=%v", victor, stakes)
	}
	if next[0].String() != "{W to play   54; Stakes: 2, W canNOT dbl, r can dbl; 1: 2: 3: 4: 5:rr 6: 7: 8: 9: 10: 11: 12: 13: 14: 15: 16:WWWWWWWWWW 17:rrrrrrrr 18: 19: 20: 21: 22: 23: 24:, 5 W off, 5 r off}" {
		t.Errorf("result=%v", next[0])
	}
}

func TestTakeTurnWhiteAcceptsQuadruple(t *testing.T) {
	rand.Seed(37)
	board := New(true)
	board.Stakes = 2
	board.WhiteCanDouble = false
	board.RedCanDouble = true
	board.Roller = White
	board.Roll = Roll{6, 6, 6, 6}
	board.Pips = Points28{}
	board.Pips[17].Reset(10, Red)
	board.Pips[16].Reset(10, White)
	board.Pips[BorneOffWhitePip].Reset(5, White)
	board.Pips[BorneOffRedPip].Reset(5, Red)
	assertValidity(board, t)
	next := board.LegalContinuations()
	victor, stakes, _ := next[0].TakeTurn(
		func(_ *Board) bool { return true },
		func(_ *Board) bool { return true })
	if victor != NoChecker || stakes != 0 {
		t.Errorf("victor=%v stakes=%v", victor, stakes)
	}
	if next[0].String() != "{r to play   54; Stakes: 4, W can dbl, r canNOT dbl; 1: 2: 3: 4: 5: 6: 7: 8: 9: 10: 11: 12: 13: 14: 15: 16:WWWWWW 17:rrrrrrrrrr 18: 19: 20: 21: 22:WWWW 23: 24:, 5 W off, 5 r off}" {
		t.Errorf("result=%v", next[0])
	}
}

func TestTakeTurnGammon(t *testing.T) {
	board := New(true)
	board.Stakes = 2
	board.WhiteCanDouble = false
	board.Roller = White
	board.Roll = Roll{6, 6, 6, 6}
	board.Pips = Points28{}
	board.Pips[1].Reset(15, Red)
	board.Pips[19].Reset(1, White)
	board.Pips[BorneOffWhitePip].Reset(14, White)
	assertValidity(board, t)
	next := board.LegalContinuations()
	if len(next) != 1 {
		t.Fatalf("TestLegalContinuations has this guy... %v", next)
	}
	victor, stakes, _ := next[0].TakeTurn(nil, nil)
	if victor != White || stakes != 4 {
		t.Errorf("victor=%v stakes=%v", victor, stakes)
	}
}

func TestTakeTurnBackgammonOnBar(t *testing.T) {
	board := New(true)
	board.Roller = Red
	board.Stakes = 128
	board.Roll = Roll{2, 1}
	board.Pips = Points28{}
	board.Pips[1].Reset(2, Red)
	board.Pips[BorneOffRedPip].Reset(13, Red)
	board.Pips[7].Reset(14, White)
	board.Pips[BarWhitePip].Reset(1, White)
	assertValidity(board, t)
	next := board.LegalContinuations()
	if len(next) != 1 {
		t.Fatalf("TestLegalContinuations has this guy... %v", next)
	}
	victor, stakes, _ := next[0].TakeTurn(nil, nil)
	if victor != Red || stakes != 128*3 {
		t.Errorf("victor=%v stakes=%v", victor, stakes)
	}
}

func TestTakeTurnBackgammonEmptyBar(t *testing.T) {
	board := New(true)
	board.Roller = Red
	board.Stakes = 128
	board.Roll = Roll{2, 1}
	board.Pips = Points28{}
	board.Pips[1].Reset(2, Red)
	board.Pips[BorneOffRedPip].Reset(13, Red)
	board.Pips[7].Reset(14, White)
	board.Pips[6].Reset(1, White)
	assertValidity(board, t)
	next := board.LegalContinuations()
	if len(next) != 1 {
		t.Fatalf("TestLegalContinuations has this guy... %v", next)
	}
	victor, stakes, _ := next[0].TakeTurn(nil, nil)
	if victor != Red || stakes != 128*3 {
		t.Errorf("victor=%v stakes=%v", victor, stakes)
	}
}

func TestTakeTurnDoubling(t *testing.T) {
	board := New(true)
	board.Roller = Red
	board.Stakes = 128
	board.Roll = Roll{2, 1}
	board.Pips = Points28{}
	board.Pips[1].Reset(2, Red)
	board.Pips[BarRedPip].Reset(1, Red)
	board.Pips[BorneOffRedPip].Reset(12, Red)
	board.Pips[7].Reset(14, White)
	board.Pips[6].Reset(1, White)
	assertValidity(board, t)
	next := board.LegalContinuations()
	if len(next) != 1 {
		t.Fatalf("TestLegalContinuations has this guy... %v", next)
	}
	startingBoard := *next[0]
	{
		next0 := startingBoard
		victor, stakes, _ := next0.TakeTurn(
			func(_ *Board) bool {
				return true
			},
			func(_ *Board) bool {
				return false
			})
		if victor != White || stakes != 128 {
			t.Errorf("victor=%v stakes=%v", victor, stakes)
		}
	}
	{
		next0 := startingBoard
		victor, stakes, _ := next0.TakeTurn(
			func(_ *Board) bool {
				return true
			},
			func(_ *Board) bool {
				return true
			})
		if victor != NoChecker || stakes != 0 {
			t.Errorf("victor=%v stakes=%v", victor, stakes)
		}
		if next0.Stakes != 256 || next0.Roller != White {
			t.Errorf("hmmm %v", next0)
		}
	}
	{
		rand.Seed(37)
		next0 := startingBoard
		victor, stakes, _ := next0.TakeTurn(
			func(_ *Board) bool {
				return false
			},
			nil)
		if victor != NoChecker || stakes != 0 {
			t.Errorf("victor=%v stakes=%v", victor, stakes)
		}
		expectedRoll := Roll{4, 1}
		if next0.Stakes != 128 || next0.Roller != White || len(next0.RollUsed.Dice()) > 0 || next0.Roll != expectedRoll {
			t.Errorf("hmmm %v", next0)
		}
	}
}

func TestPlayGame(t *testing.T) {
	type example struct {
		seed         int64
		victor       Checker
		stakes       int
		numBoard     int
		lastLog      string
		chooser      Chooser
		logger       func(interface{}, *Board)
		offerDouble  func(*Board) bool
		acceptDouble func(*Board) bool
	}
	examples := [...]example{
		example{
			36,
			White,
			3,
			64,
			"{W after playing   51; !dbl; 1:rrrrrrrrrrrr 2:r 3: 4: 5: 6: 7: 8: 9: 10: 11: 12: 13:r 14: 15: 16: 17: 18: 19: 20:r 21: 22: 23: 24:, 15 W off, Score{Goal:0,W:3,r:0,Crawford on,inactive}}",
			func(s []*Board) []AnalyzedBoard {
				return []AnalyzedBoard{AnalyzedBoard{Board: s[0]}}
			},
			func(state interface{}, b *Board) {
				if iv := b.Invalidity(IgnoreRollValidity); iv != "" {
					t.Fatalf("invalidity=%v", iv)
				}
				slicePtr := state.(*[]string)
				*slicePtr = append(*slicePtr, b.String())
			},
			nil,
			nil},

		example{
			35,
			White,
			2,
			176,
			"{W to play  666 after playing    6; !dbl; 1:rrrrrrrrrrr 2:r 3:r 4: 5: 6:r 7:r 8: 9: 10: 11: 12: 13: 14: 15: 16: 17: 18: 19: 20: 21: 22: 23: 24:, 15 W off, Score{Goal:0,W:2,r:0,Crawford on,inactive}}", // TODO(chandler37): Really? Study the entire game.
			func(s []*Board) []AnalyzedBoard {
				return []AnalyzedBoard{AnalyzedBoard{Board: s[rand.Intn(len(s))]}}
			},
			func(state interface{}, b *Board) {
				if iv := b.Invalidity(IgnoreRollValidity); iv != "" {
					t.Fatalf("invalidity=%v", iv)
				}
				slicePtr := state.(*[]string)
				*slicePtr = append(*slicePtr, b.String())
			},
			nil,
			nil},

		example{
			34,
			Red,
			1,
			56,
			"{r after playing 4444; !dbl; 1: 2: 3: 4: 5: 6: 7: 8: 9: 10: 11: 12: 13: 14: 15: 16: 17: 18: 19: 20:W 21: 22:WWWW 23:WWWWWWW 24:WW, 1 W off, 15 r off, Score{Goal:0,W:0,r:1,Crawford on,inactive}}",
			func(s []*Board) []AnalyzedBoard {
				return []AnalyzedBoard{AnalyzedBoard{Board: s[rand.Intn(len(s))]}}
			},
			func(state interface{}, b *Board) {
				if iv := b.Invalidity(IgnoreRollValidity); iv != "" {
					t.Fatalf("invalidity=%v", iv)
				}
				slicePtr := state.(*[]string)
				*slicePtr = append(*slicePtr, b.String())
			},
			nil,
			nil},

		example{
			37373737,
			White,
			3,
			111,
			"{W after playing 1111; !dbl; 1:rrrr 2:rrr 3:rr 4: 5: 6: 7: 8: 9: 10: 11: 12: 13:r 14: 15:r 16:r 17:r 18: 19:r 20: 21:r 22: 23: 24:, 15 W off, Score{Goal:0,W:3,r:0,Crawford on,inactive}}", // TODO(chandler37): Really? Study the entire game.
			func(s []*Board) []AnalyzedBoard {
				return []AnalyzedBoard{AnalyzedBoard{Board: s[rand.Intn(len(s))]}}
			},
			func(state interface{}, b *Board) {
				if iv := b.Invalidity(IgnoreRollValidity); iv != "" {
					t.Fatalf("Invalidity=%v", iv)
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
		victor, stakes, _ := New(true).PlayGame(
			&log, ex.chooser, ex.logger, ex.offerDouble, ex.acceptDouble)
		if victor != ex.victor || stakes != ex.stakes {
			t.Fatalf("exNum=%v ex=%v victor=%v stakes=%v", exNum, ex, victor, stakes)
		}
		if len(log) != ex.numBoard {
			t.Errorf("exNum=%d ex=%v len(log)=%d", exNum, ex, len(log))
		} else {
			for _, line := range log {
				t.Logf("exNum=%d: line is %v\n", exNum, line)
			}
			if l := log[len(log)-1]; l != ex.lastLog {
				t.Fatalf("exNum=%d ex=%v log[-1]=%v", exNum, ex, l)
			}
		}
	}
}

func TestLengthOfMaxPrime(t *testing.T) {
	type example struct {
		RedResult   int
		WhiteResult int
		Initializer func(*Board)
	}
	examples := [...]example{
		example{
			1,
			1,
			func(b *Board) {
			}},
		example{
			6,
			0,
			func(b *Board) {
				b.Pips = Points28{}
				for i := 1; i < 7; i++ {
					b.Pips[i].Reset(2, Red)
				}
				b.Pips[BarRedPip].Reset(3, Red)
				b.Pips[BarWhitePip].Reset(15, White)
			}},
		example{
			6,
			0,
			func(b *Board) {
				b.Pips = Points28{}
				for i := 19; i < 25; i++ {
					b.Pips[i].Reset(2, Red)
				}
				b.Pips[BarRedPip].Reset(3, Red)
				b.Pips[BarWhitePip].Reset(15, White)
			}},
		example{
			4,
			0,
			func(b *Board) {
				b.Pips = Points28{}
				for i := 1; i < 4; i++ {
					b.Pips[i].Reset(2, Red)
				}
				for i := 11; i < 15; i++ {
					b.Pips[i].Reset(2, Red)
				}
				b.Pips[BarRedPip].Reset(1, Red)
				b.Pips[BarWhitePip].Reset(15, White)
			}},
	}
	for exNum, ex := range examples {
		b := New(true)
		ex.Initializer(b)
		if iv := b.Invalidity(EnforceRollValidity); iv != "" {
			t.Errorf("invalid: %v %v %v", exNum, iv, *b)
		}
		if r := b.LengthOfMaxPrime(White); r != ex.WhiteResult {
			t.Errorf("exNum=%d: board=%v ex=%v actual W result=%v", exNum, *b, ex, r)
		}
		if r := b.LengthOfMaxPrime(Red); r != ex.RedResult {
			t.Errorf("exNum=%d: board=%v ex=%v actual r result=%v", exNum, *b, ex, r)
		}
	}
}

func TestPipCount(t *testing.T) {
	b := New(true)
	for _, player := range players {
		if x := b.PipCount(player); x != 167 {
			t.Errorf("player=%v x=%d", player, x)
		}
	}
	b.Pips[BarWhitePip].Reset(2, White)
	b.Pips[1].Reset(0, White)
	if x := b.PipCount(White); x != 167+2 {
		t.Errorf("x=%d", x)
	}

	b = New(true)
	b.Pips[BarRedPip].Reset(2, Red)
	b.Pips[6].Reset(3, Red)
	if x := b.PipCount(Red); x != 205 {
		t.Errorf("x=%d", x)
	}
	if x := b.PipCount(White); x != 167 {
		t.Errorf("x=%d", x)
	}
}

func TestBlotLiability(t *testing.T) {
	b := New(true)
	for _, player := range players {
		if x := b.BlotLiability(player); x != 0 {
			t.Errorf("player=%v x=%d", player, x)
		}
	}
	b.Pips[BarWhitePip].Reset(2, White)
	b.Pips[1].Reset(0, White)
	if x := b.BlotLiability(White); x != 0 {
		t.Errorf("x=%d", x)
	}

	b = New(true)
	b.Pips[BarRedPip].Reset(4, Red)
	b.Pips[6].Reset(1, Red)
	b.Pips[BarWhitePip].Reset(4, White)
	b.Pips[19].Reset(1, White)
	if x := b.BlotLiability(Red); x != 19 {
		t.Errorf("x=%d", x)
	}
	if x := b.BlotLiability(White); x != 19 {
		t.Errorf("x=%d", x)
	}
}

func TestRacing(t *testing.T) {
	b := New(true)
	if b.Racing() {
		t.Errorf("%v", b)
	}

	b = New(true)
	b.Pips = Points28{}
	b.Pips[13].Reset(1, White)
	b.Pips[12].Reset(1, Red)
	b.Pips[BorneOffRedPip].Reset(14, Red)
	b.Pips[BorneOffWhitePip].Reset(14, White)
	if !b.Racing() {
		t.Errorf("%v", b)
	}
}

func TestNumCheckersHome(t *testing.T) {
	b := New(true)
	if x := b.NumCheckersHome(White); x != 5 {
		t.Errorf("NumCheckersHome=%d", x)
	}

	b = New(true)
	b.Pips = Points28{}
	b.Pips[1].Reset(2, Red)
	b.Pips[6].Reset(13, Red)
	b.Pips[19].Reset(2, White)
	b.Pips[24].Reset(13, White)
	if x := b.NumCheckersHome(White); x != 15 {
		t.Errorf("NumCheckersHome=%d", x)
	}
	if x := b.NumCheckersHome(Red); x != 15 {
		t.Errorf("NumCheckersHome=%d", x)
	}
}

func TestNumPointsBlocked(t *testing.T) {
	b := New(false)
	if b.NumPointsBlocked(White) != 4 || b.NumPointsBlocked(Red) != 4 {
		t.Errorf("NumPointsBlocked is busted")
	}
}

func TestPipCountOfFarthestChecker(t *testing.T) {
	b := New(false)
	if x := b.PipCountOfFarthestChecker(White); x != 24 {
		t.Errorf("PipCountOfFarthestChecker=%d", x)
	}
	if x := b.PipCountOfFarthestChecker(Red); x != 24 {
		t.Errorf("PipCountOfFarthestChecker=%d", x)
	}
	b.Pips = Points28{}
	b.Pips[BarRedPip].Reset(15, Red)
	b.Pips[BarWhitePip].Reset(15, White)
	if x := b.PipCountOfFarthestChecker(White); x != 25 {
		t.Errorf("PipCountOfFarthestChecker=%d", x)
	}
	if x := b.PipCountOfFarthestChecker(Red); x != 25 {
		t.Errorf("PipCountOfFarthestChecker=%d", x)
	}
	b.Pips = Points28{}
	b.Pips[BorneOffWhitePip].Reset(14, White)
	b.Pips[BorneOffRedPip].Reset(14, Red)
	b.Pips[1].Reset(1, Red)
	b.Pips[24].Reset(1, White)
	if x := b.PipCountOfFarthestChecker(White); x != 1 {
		t.Errorf("PipCountOfFarthestChecker=%d", x)
	}
	if x := b.PipCountOfFarthestChecker(Red); x != 1 {
		t.Errorf("PipCountOfFarthestChecker=%d", x)
	}
}

// TODO(chandler37): Consider https://github.com/andreyvit/diff for better tests.
