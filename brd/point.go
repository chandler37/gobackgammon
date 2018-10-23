package brd

import (
	"fmt"
	"strings"
)

// A made point is a Point with two or more Checkers of the same color on
// it.
func (p *Point) MadeBy(player Checker) bool {
	return p[0] == player && p[1] == player
}

func (p Point) Equals(q Point) bool {
	for i, v := range p {
		if v != q[i] {
			return false
		}
	}
	return true
}

func (p Point) String() string {
	r := make([]string, 0, 15)
	for _, v := range p {
		if v != NoChecker {
			r = append(r, v.String())
		}
	}
	return strings.Join(r, "")
}

// Clears the Point and then places n checkers of the given color (or no color)
// on it.
func (p *Point) Reset(n int, checker Checker) {
	for i, _ := range *p {
		if i < n {
			p[i] = checker
		} else {
			p[i] = NoChecker
		}
	}
}

func (p *Point) Add(checker Checker) {
	for i, v := range *p {
		if v != checker && v != NoChecker {
			panic(fmt.Sprintf("bad point is %v", *p))
		}
		if v == NoChecker {
			p[i] = checker
			return
		}
	}
	panic("thinko")
}

func (p *Point) Subtract() {
	index := 14
	for i, v := range *p {
		if v == NoChecker {
			index = i - 1
			break
		}
	}
	if index < 0 {
		panic(fmt.Sprintf("No checkers to remove: %v", *p))
	}
	p[index] = NoChecker
}
