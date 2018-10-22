package ai

import (
	"fmt"
	"math/rand"

	"github.com/chandler37/gobackgammon/brd"
)

var debug bool = false

func StartDebugging() {
	debug = true
}

func StopDebugging() {
	debug = false
}

func randomChoice(s []*brd.Board) int {
	numNonNil := 0
	indexOfArbitraryNonNilBoard := -1
	for i, b := range s {
		if b != nil {
			numNonNil++
			indexOfArbitraryNonNilBoard = i
		}
	}
	if numNonNil < 1 {
		panic(fmt.Sprintf("The AI eliminated all %d choices", len(s)))
	}
	if numNonNil == 1 {
		return indexOfArbitraryNonNilBoard
	}
	if debug {
		fmt.Printf("DBG(random): random options are the following:\n")
	}
	rando := rand.Intn(numNonNil)
	if debug {
		for _, b := range s {
			if b != nil {
				fmt.Printf("DBG(random): %v\n", *b)
			}
		}
	}
	numNonNil = 0
	for i, b := range s {
		if b != nil {
			if numNonNil == rando {
				return i
			}
			numNonNil++
		}
	}
	panic("how?")
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

func converter(choices []brd.Board) []*brd.Board {
	result := make([]*brd.Board, 0, len(choices))
	for i, _ := range choices {
		result = append(result, &choices[i])
	}
	return result
}

func maximizer(debugLabel string, choices []*brd.Board, f func(*brd.Board) int64) []*brd.Board {
	return minmaximizer(debugLabel, choices, f, minInt64, max64)
}

func minimizer(debugLabel string, choices []*brd.Board, f func(*brd.Board) int64) []*brd.Board {
	return minmaximizer(debugLabel, choices, f, maxInt64, min64)
}

func minmaximizer(debugLabel string, choices []*brd.Board, f func(*brd.Board) int64, extremum int64, cmp func(int64, int64) int64) []*brd.Board {
	if debug {
		numNonNil := 0
		for _, c := range choices {
			if c != nil {
				numNonNil++
			}
		}
		fmt.Printf("DBG(%s): starting with %d remaining choices\n", debugLabel, numNonNil)
	}
	result := make([]*brd.Board, 0, len(choices))
	values := make([]int64, 0, len(choices))
	extremumFound := extremum
	for i, c := range choices {
		if c == nil {
			values = append(values, extremum)
		} else {
			v := f(c)
			extremumFound = cmp(extremumFound, v)
			values = append(values, v)
			if debug {
				fmt.Printf("DBG(%s): score is %d, choice is %-3d %v\n", debugLabel, v, i, *c)
			}
		}
	}
	debugging := debug
	if debug {
		numSelected := 0
		for _, v := range values {
			if v == extremumFound {
				numSelected++
			}
		}
		if numSelected <= 1 {
			debugging = false
		}
	}
	for i, v := range values {
		if v == extremumFound {
			if debugging {
				fmt.Printf("DBG(%s): %d:%v\n", debugLabel, i, *choices[i])
			}
			result = append(result, choices[i])
		} else {
			result = append(result, nil)
		}
	}
	return result
}
