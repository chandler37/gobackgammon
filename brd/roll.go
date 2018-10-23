package brd

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
)

type Die uint8 // zero: this die is not in play. 1-6: that many pips. 7+: insanity

const ZeroDie Die = 0

// You roll two dice but a doublet (e.g., <6 6>) gives you four moves (e.g., <6
// 6 6 6>). Invariants: Once you see ZeroDie, you will always see ZeroDie.
type Roll [4]Die

// <6 5> is the same as <5 6>
func (r *Roll) Equals(o Roll) bool {
	rSorted := r.Dice()
	sort.Slice(
		rSorted,
		func(i, j int) bool {
			return rSorted[i] > rSorted[j]
		})
	oSorted := o.Dice()
	sort.Slice(
		oSorted,
		func(i, j int) bool {
			return oSorted[i] > oSorted[j]
		})
	if len(rSorted) != len(oSorted) {
		return false
	}
	for i, v := range rSorted {
		if oSorted[i] != v {
			return false
		}
	}
	return true
}

func (r Roll) invalidity() string {
	u := r.UniqueDice()
	if len(u) > 2 {
		return fmt.Sprintf("too many unique dice: %v", u)
	}
	if len(u) == 2 && len(r.Dice()) > 2 {
		return fmt.Sprintf("too many effective dice: %v", r.Dice())
	}
	for i := 0; i < 2; i++ {
		if r[i] < 1 || r[i] > 6 {
			return fmt.Sprintf("die %d out of range", i)
		}
	}
	for i := 2; i < 4; i++ {
		if r[i] < 0 || r[i] > 6 {
			return fmt.Sprintf("die %d out of range", i)
		}
	}
	if r[0] < r[1] {
		return "in your test cases you must use Roll{6, 5} because Roll.New sorts the dice that way."
	}
	if r[0] == r[1] && (r[0] != r[2] || r[0] != r[3]) {
		return "in your test cases you must use Roll{6, 6, 6, 6} for a doublet"
	}
	return ""
}

// Board.RollUsed is tightly coupled to Board.Roll, so we handle both here.
func (r *Roll) New(toBeCleared *Roll) {
	for i := 0; i < len(*toBeCleared); i++ {
		(*toBeCleared)[i] = ZeroDie
	}

	x := rand.Intn(6 * 6)
	r[0] = Die((x % 6) + 1)
	r[1] = Die((x / 6) + 1)
	if r[0] < r[1] {
		// Testing is easier if we treat <5 6> and <6 5> identically. We might
		// also at some point start precomputing good moves and this will
		// improve that cache's hit rate.
		r[0], r[1] = r[1], r[0]
	}
	r[2] = ZeroDie
	r[3] = ZeroDie
	if r[0] == r[1] {
		r[2] = r[0]
		r[3] = r[0]
	}
}

func (r Roll) String() string {
	parts := []string{}
	for _, die := range r {
		if die != ZeroDie {
			parts = append(parts, fmt.Sprintf("%d", die))
		}
	}
	if len(parts) == 0 {
		return "<>"
	}
	return fmt.Sprintf("%4s", strings.Join(parts, ""))
}

// Returns a new Roll the same as r minus die. Mutates recipient to add a die.
func (r Roll) Use(die Die, recipient *Roll) Roll {
	if die < 1 || die > 6 {
		panic(fmt.Sprintf("die is %d", die))
	}
	j := 0
	result := Roll{}
	alreadyUsed := false
	for _, d := range r {
		if !alreadyUsed && d == die {
			alreadyUsed = true
			continue
		}
		if d != die || alreadyUsed {
			result[j] = d
			j++
		}
	}
	found := false
	for i, d := range *recipient {
		if d == ZeroDie {
			(*recipient)[i] = die
			found = true
			break
		}
	}
	if !found {
		panic("recipient was full")
	}
	return result
}

// Returns the unique dice.
//
// Roll{6, 6, 6, 6}.UniqueDice() => []Die{6}
func (r Roll) UniqueDice() []Die {
	presence := make(map[Die]struct{}, 6)
	for _, d := range r {
		if d == ZeroDie {
			continue
		}
		presence[d] = struct{}{}
	}
	unsorted := make([]Die, 0, len(presence))
	for k, _ := range presence {
		unsorted = append(unsorted, k)
	}
	sort.Slice(
		unsorted,
		func(i, j int) bool {
			return unsorted[i] > unsorted[j]
		})
	return unsorted
}

// Returns the dice.
//
// Roll{6, 6, 6, 6}.Dice() => []Die{6, 6, 6, 6}
// Roll{6, 1}.Dice() => []Die{6, 1}
func (r Roll) Dice() (result []Die) {
	for _, d := range r {
		if d != ZeroDie {
			result = append(result, d)
		}
	}
	return
}
