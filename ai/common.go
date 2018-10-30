package ai

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"

	"github.com/chandler37/gobackgammon/brd"
)

var debug bool = false

func StartDebugging() {
	debug = true
}

func StopDebugging() {
	debug = false
}

func shuffle(choices []brd.AnalyzedBoard) {
	if len(choices) > 1 {
		// a tie is fine:
		maximizer("randomizer", choices, func(*brd.Board) int64 { return rand.Int63() })
	}
}

const (
	minInt64 int64 = -9223372036854775808
	maxInt64 int64 = 9223372036854775807
)

func max64(i, j int64) int64 {
	if i < j {
		return j
	}
	return i
}

func min64(i, j int64) int64 {
	if i < j {
		return i
	}
	return j
}

func converter(choices []*brd.Board) []brd.AnalyzedBoard {
	result := make([]brd.AnalyzedBoard, len(choices), len(choices))
	for i, _ := range choices {
		result[i].Board = choices[i]
	}
	return result
}

func maximizer(label label, choices []brd.AnalyzedBoard, f func(*brd.Board) int64) {
	minmaximizer(label, choices, f, minInt64, max64)
}

func minimizer(label label, choices []brd.AnalyzedBoard, f func(*brd.Board) int64) {
	minmaximizer(label, choices, f, maxInt64, min64)
}

type label string

type minmaxAnalysis struct {
	RuledOut string // empty string: not ruled out yet. Human readable.
	Scores   scoreMap
}

type scoreMap map[label]int64

func (s scoreMap) String() string {
	keys := make([]label, 0, len(s))
	for k, _ := range s {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	parts := make([]string, 0, len(s))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%v=%v", k, s[k]))
	}
	return strings.Join(parts, " ")
}

func (m minmaxAnalysis) Summary() string {
	if m.RuledOut == "" {
		return "wasn't ruled out by heuristics. Details: " + m.Scores.String()
	}
	return m.RuledOut
}

func analyzedBoardIsRuledOut(a brd.AnalyzedBoard) bool {
	return a.Analysis != nil && a.Analysis.(minmaxAnalysis).RuledOut != ""
}

func minmaximizer(label label, choices []brd.AnalyzedBoard, f func(*brd.Board) int64, extremum int64, cmp func(int64, int64) int64) {
	if debug {
		numNotRuledOut := 0
		for _, c := range choices {
			if !analyzedBoardIsRuledOut(c) {
				numNotRuledOut++
			}
		}
		fmt.Printf("DBG(%s): starting with %d remaining choices\n", label, numNotRuledOut)
	}
	extremumFound := extremum
	for i, c := range choices {
		if !analyzedBoardIsRuledOut(c) {
			v := f(c.Board)
			extremumFound = cmp(extremumFound, v)
			if choices[i].Analysis == nil {
				choices[i].Analysis = minmaxAnalysis{Scores: make(scoreMap)}
			}
			choices[i].Analysis.(minmaxAnalysis).Scores[label] = v
			if debug {
				fmt.Printf("DBG(%s): score is %d, choice is %-3d %v\n", label, v, i, *c.Board)
			}
		}
	}
	debugging := debug
	if debug {
		numSelected := 0
		for _, c := range choices {
			if c.Analysis != nil && c.Analysis.(minmaxAnalysis).Scores[label] == extremumFound {
				numSelected++
			}
		}
		if numSelected <= 1 {
			debugging = false
		}
	}
	result := make([]brd.AnalyzedBoard, 0, len(choices))
	for i, c := range choices {
		if c.Analysis != nil && c.Analysis.(minmaxAnalysis).Scores[label] == extremumFound {
			if debugging {
				fmt.Printf("DBG(%s): score=%v %d:%v\n", label, extremumFound, i, *choices[i].Board)
			}
			result = append(result, c)
		}
	}
	remainder := make([]brd.AnalyzedBoard, 0, len(choices)-len(result))
	for _, c := range choices {
		if analyzedBoardIsRuledOut(c) {
			continue
		}
		if a := c.Analysis.(minmaxAnalysis); a.Scores[label] != extremumFound {
			if a.RuledOut != "" {
				panic("already ruled out: " + a.RuledOut)
			}
			a.RuledOut = fmt.Sprintf("Ruled out by %v (%v)", label, a.Scores[label])
			c.Analysis = a
			remainder = append(remainder, c)
		}
	}
	sort.SliceStable(
		remainder,
		func(i, j int) bool {
			return remainder[i].Analysis.(minmaxAnalysis).Scores[label] < remainder[j].Analysis.(minmaxAnalysis).Scores[label]
		})
	result = append(result, remainder...)
	copy(choices, result)
}

// TODO(chandler37): Add heuristics that avoid gammons, too, but aware of
// tournament play and the Jacoby rule.
func probabilityOfGettingBackgammoned(b *brd.Board) (score int64) {
	if b.MatchScore.Goal > 0 {
		otherPlayerScore := b.MatchScore.RedScore
		if b.Roller == brd.Red {
			otherPlayerScore = b.MatchScore.WhiteScore
		}
		if otherPlayerScore+1 > b.MatchScore.Goal {
			return -1 // a backgammon is the same as a single-stakes loss
		}
	}
	// The number of checkers on the bar is a constant for legal
	// continutations.
	if b.Roller == brd.White {
		for i := 1; i < 7; i++ {
			score += int64(7-i) * int64(b.Pips[i].NumWhite())
		}
		return
	}
	for i := 19; i < 25; i++ {
		score += int64(i-18) * int64(b.Pips[i].NumRed())
	}
	return
}
