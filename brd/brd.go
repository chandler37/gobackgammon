package brd

import (
	"fmt"
	"math/rand"
	"strings"
)

// Each player has 15 Checkers. A checker is either on a point, on the bar, or borne off.
type Checker uint8

const (
	NoChecker Checker = iota
	White
	Red
)

// A point, sometimes called a pip, is represented on the board by a
// triangle. Zero or more checkers, all of like color, may occupy a
// point/pip. Values are [-15, 15] where -3 means there are three White
// checkers and 15 means there are fifteen Red checkers and 0 means there are
// no checkers.
type Point int8

type Points28 [MaxPip + 1]Point

// TODO(chandler37): use sync.Pool for a Board pool and see how much it speeds
// up the highest-level benchmarks to avoid giving the garbage collector so
// much work.

// See http://www.gammonlife.com/rules/backgammon-rules.htm for a picture of a
// board with points annotated [1, 24]. The standard starting position has two
// white checkers on 1 and two red checkers on 24. We also have a notion of
// point 0, the red checkers that have borne off (during the end stage of the
// game, bearing off), and point 25, the white checkers that have borne off.
//
// A deep copy of a Board is the very same as a shallow copy. No slices, maps,
// pointers, funcs, channels, etc.
type Board struct {
	// Zero values may appear anywhere. We use [4]Die instead of []Die for
	// efficiency's sake and to make deep copying easy:
	Roll           Roll // unused
	RollUsed       Roll // used
	Roller         Checker
	Stakes         int
	MatchScore     Score // zero value means we are not doing tournament play. NB: Use SetScore()
	WhiteCanDouble bool
	RedCanDouble   bool
	Pips           Points28 // red borne off, the 24 points of the board, white borne off, red bar, white bar. Pips[1:25] are the 24 pips.
}

// TODO(chandler37): Test the AIs with a 6-prime from [6, 12) or even farther from home.

// An intelligence (a player) has to make decisions about the doubling cube and
// choose moves. A Chooser is just the part that chooses moves after the
// doubling phase.
//
// Element zero of the returned slice is the best move, element len(result)-1
// is the worst. If the analyses indicate a tie, it doesn't affect game
// play. If you want to choose randomly from tied Boards, you must shuffle them
// yourself.
//
// You are free to return a subset of the input, e.g. a single Board, the Kth:
// []AnalyzedBoard{AnalyzedBoard{Board:input[K]}}. You must return at least one.
//
// Your output must not contain nil Board pointers. You must not mutate the
// Boards pointed to by the input.
type Chooser func([]*Board) []AnalyzedBoard

type AnalyzedBoard struct {
	Board    *Board
	Analysis Analysis // optional
}

type Analysis interface {
	Summary() string // for human consumption
}

func (b AnalyzedBoard) String() string {
	if b.Analysis == nil {
		return b.Board.String()
	}
	return fmt.Sprintf("%v (%v)", b.Board, b.Analysis.Summary())
}

// TODO(chandler37): create a high-level API for changing the RNG

// Read-only method using a pointer receiver to be performant. The function
// that are parameters use pointers and slices for efficiency's sake but must
// not mutate their arguments.
//
// Invariant: the game is not yet over.
//
// Calls logger, if non-nil, for each Board, including b at the start.
//
// Does not attempt doubling for the starting board.
//
// Returns the victor and the stakes (which is 1, 2 (gammon), or 3 (backgammon)
// multiplied by the final Board's Stakes) and the match score.
//
// TODO(chandler37): Perhaps also wrap this up with a "you cannot cheat or
// accidentally mess things up" version that never lets you mutate state (i.e.,
// deep copies the []*Board and never gives a *Board)?
func (b *Board) PlayGame(logState interface{}, chooser Chooser, logger func(interface{}, *Board), offerDouble, acceptDouble func(*Board) bool) (Checker, int, Score) {
	if logger != nil {
		logger(logState, b)
	}
	currentBoard := b
	for {
		candidates := currentBoard.LegalContinuations()
		analyzedCandidates := chooser(candidates)
		if len(analyzedCandidates) > len(candidates) {
			panic("madness")
		}
		if len(analyzedCandidates) == 0 {
			currentBoard = candidates[0]
		} else {
			currentBoard = analyzedCandidates[0].Board
		}
		victor, stakes, score := currentBoard.TakeTurn(offerDouble, acceptDouble)
		logger(logState, currentBoard)
		if victor != NoChecker {
			return victor, stakes, score
		}
	}
}

// Flips the Roller, offers a double, rolls new dice, alters the MatchScore.
//
// offerDouble and acceptDouble may be nil.
//
// Invariant: the receiver was returned by LegalContinuations()
//
// Mutates the receiver.
func (b *Board) TakeTurn(offerDouble, acceptDouble func(*Board) bool) (victor Checker, stakes int, score Score) {
	if victor, stakes = b.victor(); victor != NoChecker {
		b.MatchScore.Update(victor, stakes)
		score = b.MatchScore
		return
	}
	b.Roller = b.Roller.OtherColor()
	b.Roll = Roll{}
	if (b.Roller == Red && b.RedCanDouble) || (b.Roller == White && b.WhiteCanDouble) {
		if offerDouble != nil && offerDouble(b) {
			if acceptDouble(b) {
				b.Stakes *= 2
				if b.Roller == White {
					b.WhiteCanDouble = false
					b.RedCanDouble = true
				} else {
					b.RedCanDouble = false
					b.WhiteCanDouble = true
				}
			} else {
				victor = b.Roller
				stakes = b.Stakes
				return
			}
		}
	}
	b.Roll.New(&b.RollUsed)
	return
}

func (b *Board) NumPointsBlocked(player Checker) (result int) {
	for i := 1; i < 25; i++ {
		if b.Pips[i].MadeBy(player) {
			result++
		}
	}
	return
}

// A 4-prime, e.g., is four consecutive made Points.
func (b *Board) LengthOfMaxPrime(player Checker) (result int) {
	k := 0
	for i := 1; i < 25; i++ {
		if b.Pips[i].MadeBy(player) {
			k++
		} else {
			result = max(k, result)
			k = 0
		}
	}
	result = max(k, result)
	return
}

func (b *Board) NumCheckersHome(player Checker) (result int) {
	home := b.Pips[19:25]
	if player == Red {
		home = b.Pips[1:7]
	}
	for _, p := range home {
		if player == Red {
			if p > 0 {
				result += int(p)
			}
		} else {
			if p < 0 {
				result += int(-p)
			}
		}
	}
	return
}

func (b *Board) PipCount(player Checker) (result int) {
	if player == White {
		result += 25 * b.Pips[BarWhitePip].NumCheckers()
		for i := 1; i < 25; i++ {
			result += (25 - i) * b.Pips[i].NumWhite()
		}
		return
	}
	result += 25 * b.Pips[BarRedPip].NumCheckers()
	for i := 1; i < 25; i++ {
		result += i * b.Pips[i].NumRed()
	}
	return
}

func (b *Board) PipCountOfFarthestChecker(player Checker) int {
	if player == White {
		extremeWhite := -1
		if b.Pips[BarWhitePip].NumCheckers() > 0 {
			extremeWhite = 0
		} else {
			for i := 1; i < 25; i++ {
				if b.Pips[i].NumWhite() > 0 {
					extremeWhite = i
					break
				}
			}
		}
		return 25 - extremeWhite
	}
	extremeRed := -1
	if b.Pips[BarRedPip].NumCheckers() > 0 {
		extremeRed = 25
	} else {
		for i := 24; i > 0; i-- {
			if b.Pips[i].NumRed() > 0 {
				extremeRed = i
				break
			}
		}
	}
	return extremeRed
}

// A "race" is when it is impossible for either player to hit the other.
func (b *Board) Racing() bool {
	extremeWhite := -1
	if b.Pips[BarWhitePip].NumCheckers() > 0 {
		extremeWhite = 0
	} else {
		for i := 1; i < 25; i++ {
			if b.Pips[i].NumWhite() > 0 {
				extremeWhite = i
				break
			}
		}
	}

	extremeRed := -1
	if b.Pips[BarRedPip].NumCheckers() > 0 {
		extremeRed = 25
	} else {
		for i := 24; i > 0; i-- {
			if b.Pips[i].NumRed() > 0 {
				extremeRed = i
				break
			}
		}
	}

	return extremeRed < extremeWhite
}

func (b *Board) BlotLiability(player Checker) (result int) {
	fn := func(index int) int {
		return index
	}
	if player == Red {
		fn = func(index int) int {
			return 25 - index
		}
	}
	for i := 1; i < 25; i++ {
		if b.Pips[i].Num(player) == 1 {
			result += fn(i)
		}
	}
	return
}

// Read-only method using a pointer receiver to be performant.
//
// Enumerates all legal Board continuations without duplicates. The result is
// guaranteed to be non-empty (sometimes it's just []*Board{b}).
//
// The resulting Boards have the same Roller and the same dice (though they may
// be shifted from Roll to RollUsed). You must call TakeTurn() next.
func (b *Board) LegalContinuations() []*Board {
	candidates := b.quasiLegalContinuations()
	if len(candidates) < 1 {
		panic("the no-op isn't there")
	}
	maxCandidates := make([]*Board, 0, len(candidates))
	maxDiceUsed := 0
	for _, c := range candidates {
		maxDiceUsed = max(len(c.RollUsed.Dice()), maxDiceUsed)
	}
	for _, c := range candidates {
		if len(c.RollUsed.Dice()) == maxDiceUsed {
			maxCandidates = append(maxCandidates, c)
		}
	}
	// But it's not legal to take just a <3> if you can take just a <6>. We
	// have more work to do. (What if you can take both the <3> and the <6>?
	// Then you must, but the above loops would weed out all possibilities
	// except those that use both.)
	arbitraryCandidate := maxCandidates[0]
	if len(arbitraryCandidate.RollUsed.Dice()) != 1 {
		return maxCandidates
	}
	results := make([]*Board, 0, len(maxCandidates))
	var maxDieUsed Die
	for _, c := range maxCandidates {
		if dieUsed := c.RollUsed.Dice()[0]; dieUsed == ZeroDie {
			panic(b.String())
		} else {
			maxDieUsed = maxDie(maxDieUsed, dieUsed)
		}
	}
	for _, c := range maxCandidates {
		if maxDieUsed == c.RollUsed.Dice()[0] {
			results = append(results, c)
		}
	}
	return results
}

func (x Points28) Equals(y Points28) bool {
	for i, v := range x {
		if !v.Equals(y[i]) {
			return false
		}
	}
	return true
}

func (b *Board) Equals(o Board) bool {
	if !b.Pips.Equals(o.Pips) {
		return false
	}
	if !b.MatchScore.Equals(o.MatchScore) {
		return false
	}
	if b.Roller != o.Roller {
		return false
	}
	if !b.Roll.Equals(o.Roll) {
		return false
	}
	if !b.RollUsed.Equals(o.RollUsed) {
		return false
	}
	if b.Stakes != o.Stakes {
		return false
	}
	if b.WhiteCanDouble != o.WhiteCanDouble {
		return false
	}
	if b.RedCanDouble != o.RedCanDouble {
		return false
	}
	return true
}

func New(paranoid bool) *Board {
	board := Board{Stakes: 1, WhiteCanDouble: true, RedCanDouble: true}
	board.Pips[1].Reset(2, White)
	board.Pips[24].Reset(2, Red)
	board.Pips[6].Reset(5, Red)
	board.Pips[19].Reset(5, White)
	board.Pips[8].Reset(3, Red)
	board.Pips[17].Reset(3, White)
	board.Pips[12].Reset(5, White)
	board.Pips[13].Reset(5, Red)

	board.Roller = players[rand.Intn(2)]
	for {
		board.Roll.New(&board.RollUsed)
		if board.Roll[0] != board.Roll[1] {
			break
		}
	}
	if paranoid {
		if v := board.Invalidity(EnforceRollValidity); v != "" {
			panic(v)
		}
	}
	return &board
}

const (
	BorneOffWhitePip = 0
	BorneOffRedPip   = 25
	BarWhitePip      = 26
	BarRedPip        = 27
	MaxPip           = BarRedPip
)

// useful for debugging the tests:
const paddedStrings = false

func (b Board) String() string {
	score := ""
	if !b.MatchScore.Equals(Score{}) {
		score = ", " + b.MatchScore.String()
	}
	barWhite := ""
	l := []string{}
	for n := 0; n < b.Pips[BarWhitePip].NumCheckers(); n++ {
		l = append(l, fmt.Sprintf("%v", White))
	}
	if len(l) > 0 {
		barWhite = fmt.Sprintf(", %v on bar", strings.Join(l, ""))
	}
	barRed := ""
	l = []string{}
	for n := 0; n < b.Pips[BarRedPip].NumCheckers(); n++ {
		l = append(l, fmt.Sprintf("%v", Red))
	}
	if len(l) > 0 {
		barRed = fmt.Sprintf(", %v on bar", strings.Join(l, ""))
	}
	borneOffWhite := ""
	num := b.Pips[BorneOffWhitePip].NumCheckers()
	if num > 0 {
		borneOffWhite = fmt.Sprintf(", %d %v off", num, White)
	}
	borneOffRed := ""
	num = b.Pips[BorneOffRedPip].NumCheckers()
	if num > 0 {
		borneOffRed = fmt.Sprintf(", %d %v off", num, Red)
	}
	prettyPips := make([]string, 0, 24)
	for i := 1; i < 25; i++ {
		if paddedStrings {
			prettyPips = append(prettyPips, fmt.Sprintf("%02d:%-9v", i, b.Pips[i]))
		} else {
			prettyPips = append(prettyPips, fmt.Sprintf("%d:%v", i, b.Pips[i]))
		}
	}
	whiteDouble := "NOT"
	if b.WhiteCanDouble {
		whiteDouble = ""
	}
	redDouble := "NOT"
	if b.RedCanDouble {
		redDouble = ""
	}
	stakes := fmt.Sprintf(
		"Stakes: %d, %v can%s dbl, %v can%s dbl",
		b.Stakes, White, whiteDouble, Red, redDouble)
	if b.Stakes == 1 && b.WhiteCanDouble && b.RedCanDouble {
		stakes = "!dbl"
	}
	usedRoll := ""
	if len(b.RollUsed.Dice()) > 0 {
		usedRoll = fmt.Sprintf(" after playing %v", b.RollUsed)
	}
	toPlay := ""
	if len(b.Roll.Dice()) > 0 {
		toPlay = fmt.Sprintf(" to play %v", b.Roll)
	}
	return fmt.Sprintf(
		"{%v%s%s; %s; %v%s%s%s%s%s}",
		b.Roller, toPlay, usedRoll, stakes, strings.Join(prettyPips, " "), barWhite,
		barRed, borneOffWhite, borneOffRed, score)
}

const (
	IgnoreRollValidity  = true
	EnforceRollValidity = !IgnoreRollValidity
)

func (b Board) Invalidity(ignoreRoll bool) string {
	if !ignoreRoll {
		if i := b.Roll.invalidity(); i != "" {
			return fmt.Sprintf("Invalid roll: %s", i)
		}
	}
	if b.Roller != Red && b.Roller != White {
		return "bad Roller"
	}
	numWhite, numRed := b.Pips[BarWhitePip].NumWhite(), b.Pips[BarRedPip].NumRed()
	if b.Pips[BarWhitePip].NumRed() > 0 {
		return "Red on BarWhite"
	}
	if b.Pips[BarRedPip].NumWhite() > 0 {
		return "White on BarRed"
	}
	if b.Pips[BorneOffWhitePip].NumRed() > 0 {
		return "Red on BorneOffWhitePip"
	}
	if b.Pips[BorneOffRedPip].NumWhite() > 0 {
		return "Red on BorneOffRedPip"
	}
	numWhite += b.Pips[BorneOffWhitePip].NumWhite()
	numRed += b.Pips[BorneOffRedPip].NumRed()
	for pointNumber := 1; pointNumber < 25; pointNumber++ {
		if b.Pips[pointNumber] < -15 || b.Pips[pointNumber] > 15 {
			return "out of range [-15, 15]"
		}
		numWhite += b.Pips[pointNumber].NumWhite()
		numRed += b.Pips[pointNumber].NumRed()
	}
	if numWhite != 15 {
		return fmt.Sprintf("%d White checkers found, not 15", numWhite)
	}
	if numRed != 15 {
		return fmt.Sprintf("%d Red checkers found, not 15", numRed)
	}
	return ""
}

func max(i, j int) int {
	if i < j {
		return j
	}
	return i
}

// assumes victory for b.Roller.
func (b *Board) victorMultiplier() int {
	opponentBar := BarWhitePip
	opponentBorne := BorneOffWhitePip
	homeStart, homeEnd := 1, 6
	if b.Roller == White {
		opponentBar = BarRedPip
		opponentBorne = BorneOffRedPip
		homeStart, homeEnd = 19, 24
	}
	if b.Pips[opponentBar].NumCheckers() > 0 {
		return 3
	}
	for x := homeStart; x <= homeEnd; x++ {
		if b.Pips[x].NumCheckers() > 0 {
			return 3
		}
	}
	if b.Pips[opponentBorne].NumCheckers() == 0 {
		return 2
	}
	return 1
}

func (b *Board) victor() (victor Checker, stakes int) {
	borne := BorneOffRedPip
	if b.Roller == White {
		borne = BorneOffWhitePip
	}
	if b.Pips[borne].NumCheckers() == 15 {
		victor = b.Roller
		stakes = b.victorMultiplier() * b.Stakes
		return
	}
	return
}

var players = [...]Checker{White, Red}

func (c Checker) OtherColor() Checker {
	switch c {
	case Red:
		return White
	case White:
		return Red
	default:
		return NoChecker
	}
}

func (b *Board) pipIsBlockedByOpponent(i int) bool {
	opponent := b.Roller.OtherColor()
	return b.Pips[i].MadeBy(opponent)
}

// Returns a Board or nil depending on whether or not that point was open.
func (b *Board) comeOffTheBar(die Die) *Board {
	// At the start, Pips[1] is Point{White, White}. If b.Roller is White, then
	// we come in on the die Point. Else the 25-die point.
	i := int(die)
	barPip := BarWhitePip
	otherPlayersBar := BarRedPip
	switch b.Roller {
	case Red:
		i = 25 - int(die)
		barPip = BarRedPip
		otherPlayersBar = BarWhitePip
	case White:
	default:
		panic("bad b.Roller")
	}
	if b.pipIsBlockedByOpponent(i) {
		return nil
	}
	result := *b
	result.Roll = result.Roll.Use(die, &result.RollUsed)
	if other := b.Roller.OtherColor(); result.Pips[i].Num(other) > 0 {
		result.Pips[i].Reset(0, White)
		err := result.Pips[otherPlayersBar].Add(other)
		if err != nil {
			panic(fmt.Sprintf("b=%v b.Pips=%q err=%v other=%v otherPlayersBar=%v", *b, b.Pips, err, other, otherPlayersBar))
		}
	}
	result.Pips[i].Add(b.Roller)
	result.Pips[barPip].Subtract()
	return &result
}

// possibilities will never be empty. It will sometimes be []*Board{b} e.g. if
// there's nothing on the bar or if there's something on the bar that is
// blocked from coming in. It will be multiple Boards if a Checker on the bar
// can come in on multiple Points.
func (b *Board) continuationsOffTheBar() (possibilities []*Board) {
	// This is recursive, and the base case for our recursion is if (1)
	// b.Roller has none on the bar or (2) the b.Roll is exhausted.
	numOnBar := b.numCheckersRollerHasOnTheBar()
	if numOnBar > 0 {
		for _, die := range b.Roll.UniqueDice() {
			if next := b.comeOffTheBar(die); next != nil {
				possibilities = append(possibilities, next.continuationsOffTheBar()...)
			}
		}
	}
	if len(possibilities) == 0 {
		possibilities = []*Board{b}
	}
	return
}

func (b *Board) numCheckersRollerHasOnTheBar() int {
	barPip := BarWhitePip
	if b.Roller == Red {
		barPip = BarRedPip
	}
	return b.Pips[barPip].NumCheckers()
}

func (b *Board) canBearOff() bool {
	if b.numCheckersRollerHasOnTheBar() > 0 {
		return false
	}
	if b.Roller == White {
		for i := 1; i < 19; i++ {
			if b.Pips[i].NumWhite() > 0 {
				return false
			}
		}
		return true
	}
	for i := 7; i < 25; i++ {
		if b.Pips[i].NumRed() > 0 {
			return false
		}
	}
	return true
}

// targetPip is undefined unless can is true
func (b *Board) canMoveChecker(startPipIndex int, die Die) (targetPip int, can bool) {
	if b.Pips[startPipIndex].Num(b.Roller) < 1 || startPipIndex < 1 || startPipIndex > 24 {
		panic("@the disco")
	}
	if b.Roller == White {
		targetPip = startPipIndex + int(die)
		if startPipIndex >= 19 && targetPip > 24 {
			exact := targetPip == 25
			goodEnough := true
			if targetPip != 25 {
				for i := 19; i < startPipIndex; i++ {
					if b.Pips[i].Num(b.Roller) > 0 {
						goodEnough = false
						break
					}
				}
			}
			if exact || goodEnough {
				can = b.canBearOff()
				targetPip = BorneOffWhitePip
			}
			return
		}
		can = !b.pipIsBlockedByOpponent(targetPip)
		return
	}
	targetPip = startPipIndex - int(die)
	if startPipIndex <= 6 && targetPip < 1 {
		exact := targetPip == 0
		goodEnough := true
		if targetPip != 0 {
			for i := 6; i > startPipIndex; i-- {
				if b.Pips[i].Num(b.Roller) > 0 {
					goodEnough = false
					break
				}
			}
		}
		if exact || goodEnough {
			can = b.canBearOff()
			targetPip = BorneOffRedPip
		}
		return
	}
	can = !b.pipIsBlockedByOpponent(targetPip)
	return
}

// Invariant: len(b.Roll.Dice()) > 0 && b.numCheckersRollerHasOnTheBar() == 0
func (b *Board) quasiLegalPostBarContinuations() (continuations []*Board) {
	remainingDice := b.Roll.Dice()
	if len(remainingDice) == 0 || b.numCheckersRollerHasOnTheBar() > 0 {
		return []*Board{b}
	}
	// Imagine a starting board with White rolling <6 5>. We must examine both
	// <5 6> and <6 5> to see the possibility of moving from point 1 to point
	// 12. This does so.
	for _, die := range remainingDice {
		for i := 1; i < 25; i++ {
			if b.Pips[i].Num(b.Roller) > 0 {
				if targetPip, can := b.canMoveChecker(i, die); can {
					next := *b
					next.Pips[i].Subtract()
					if other := b.Roller.OtherColor(); next.Pips[targetPip].Num(other) > 0 {
						next.Pips[targetPip].Subtract()
						bar := BarRedPip
						if other == White {
							bar = BarWhitePip
						}
						next.Pips[bar].Add(other)
					}
					next.Pips[targetPip].Add(b.Roller)
					next.Roll = next.Roll.Use(die, &next.RollUsed)
					continuations = append(continuations, next.quasiLegalPostBarContinuations()...)
				}
			}
		}
	}
	if len(continuations) == 0 {
		continuations = append(continuations, b)
	} else {
		continuations = uniqueContinuations(continuations)
	}
	return
}

// to ease testing, this must be stable, i.e., not rearranging things
func uniqueContinuations(continuations []*Board) []*Board {
	result := make([]*Board, 0, len(continuations))
	for _, m := range continuations {
		unique := true
		for _, r := range result {
			if m.Equals(*r) {
				unique = false
				break
			}
		}
		if unique {
			result = append(result, m)
		}
	}
	if len(result) == 0 {
		panic(fmt.Sprintf("input had %d elements", len(continuations)))
	}
	return result
}

// For a <6 3> this returns Boards where we took just the <6>, just the <3>,
// and also, if possible, where we took both. The legal continuations are the ones
// where we took both, or, if there are no such continuations, the boards where we took
// the <6>. Yes, if you can only take one, you must take the larger. If you can
// take four, you must take four, but this may return boards where we took only
// one, two, or three. (You must take the max possible. If you can take three
// but not four, you must. If you can take two, you must. if you can take one,
// you must.)
func (b *Board) quasiLegalContinuations() []*Board {
	barContinuations := b.continuationsOffTheBar()
	continuations := []*Board{}
	for _, next := range barContinuations {
		continuations = append(continuations, next.quasiLegalPostBarContinuations()...)
	}
	return uniqueContinuations(continuations)
}

func maxDie(i, j Die) Die {
	if i < j {
		return j
	}
	return i
}

func (c Checker) String() string {
	if c == White {
		return "W"
	}
	if c == Red {
		return "r"
	}
	return "0"
}

func (b *Board) SetScore(score Score) {
	b.MatchScore = score
	if score.CrawfordRuleAppliesNextGame() {
		b.WhiteCanDouble = false
		b.RedCanDouble = false
		b.MatchScore.AlreadyPlayedCrawfordGame = true
	}
}
