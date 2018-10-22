package brd

import (
	"fmt"
)

// A match to 5 using the Crawford rule would be Score{Goal: 5}. The Crawford
// rule is a standard so our zero value uses it.
type Score struct {
	WhiteScore                int
	RedScore                  int
	NoCrawfordRule            bool
	AlreadyPlayedCrawfordGame bool // TODO(chandler37): a test case where the crawford game is the last game, and another where the crawford game is not the last.
	Goal                      int  // zero means we are not playing a match, just looking to maximize points.
}

func (s *Score) Equals(o Score) bool {
	if s.WhiteScore != o.WhiteScore {
		return false
	}
	if s.RedScore != o.RedScore {
		return false
	}
	if s.NoCrawfordRule != o.NoCrawfordRule {
		return false
	}
	if s.AlreadyPlayedCrawfordGame != o.AlreadyPlayedCrawfordGame {
		return false
	}
	if s.Goal != o.Goal {
		return false
	}
	return true
}

func (s Score) String() string {
	craw := ""
	if s.NoCrawfordRule {
		craw = fmt.Sprintf("off")
	} else {
		x := "inactive"
		if s.AlreadyPlayedCrawfordGame {
			x = "dormant"
		}
		craw = fmt.Sprintf("on,%s", x)
	}
	return fmt.Sprintf(
		"Score{Goal:%d,%v:%d,%v:%d,Crawford %s}",
		s.Goal, White, s.WhiteScore, Red, s.RedScore, craw)
}

func (s *Score) Update(victor Checker, stakes int) {
	if (victor != Red && victor != White) || stakes < 1 {
		panic("bad victor")
	}
	if victor == White {
		s.WhiteScore += stakes
	} else {
		s.RedScore += stakes
	}
}

func (s *Score) CrawfordRuleAppliesNextGame() bool {
	if s.NoCrawfordRule || s.AlreadyPlayedCrawfordGame {
		return false
	}
	return s.RedScore+1 == s.Goal || s.WhiteScore+1 == s.Goal
}
