package brd

import (
	"fmt"
	"strings"
)

func (p Point) NumCheckers() int {
	if p < 0 {
		return int(-p)
	}
	return int(p)
}

func (p Point) NumWhite() int {
	if p < 0 {
		return int(-p)
	}
	return 0
}

func (p Point) NumRed() int {
	if p > 0 {
		return int(p)
	}
	return 0
}

func (p Point) Num(player Checker) int {
	if player == Red {
		return p.NumRed()
	}
	if player != White {
		panic("bad player")
	}
	return p.NumWhite()
}

// A made point is a Point with two or more Checkers of the same color on
// it.
func (p Point) MadeBy(player Checker) bool {
	return p.Num(player) >= 2
}

func (p Point) Equals(q Point) bool {
	return p == q
}

func (p Point) String() string {
	r := make([]string, 0, 15)
	for n := 0; n < p.NumWhite(); n++ {
		r = append(r, White.String())
	}
	for n := 0; n < p.NumRed(); n++ {
		r = append(r, Red.String())
	}
	return strings.Join(r, "")
}

// Clears the Point and then places n checkers of the given color (or no color)
// on it.
func (p *Point) Reset(n int, checker Checker) {
	if n < 0 || n > 15 {
		panic("bad n")
	}
	if checker == White {
		*p = Point(-n)
		return
	}
	*p = Point(n)
}

func (p *Point) Add(checker Checker) error {
	if checker == White {
		if p.NumRed() > 0 {
			return fmt.Errorf("when adding White you must zero first: %d", *p)
		}
		*p -= 1
		return nil
	}
	if p.NumWhite() > 0 {
		return fmt.Errorf("when adding Red you must zero first: %d", *p)
	}
	*p += 1
	return nil
}

func (p *Point) Subtract() {
	if *p == 0 {
		panic("cannot subtract an empty point")
	}
	if *p < 0 {
		*p += 1
		return
	}
	*p -= 1
}

// Returns a Point with n checkers on it of the given color.
func NewPoint(n int, color Checker) Point {
	if n < 0 || n > 15 {
		panic("bad n")
	}
	if color != White && color != Red {
		panic("bad color")
	}
	if color == White {
		return Point(-n)
	}
	return Point(n)
}
