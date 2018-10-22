// Package ai represents a backgammon player by a function with the signature
// brd.Chooser (i.e., func([]brd.Board) int).
//
// You pass a brd.Chooser to brd.PlayGame(). You must seed math/rand
// appropriately.
//
// TODO(chandler37): Implement an AI that chooses not just its continuation
// brd.Board but also its brd.Roll. If you leave it a blot, it's very likely to
// choose the next roll that hits it and makes a point on it. If you don't
// leave it a blot, it'll probably roll high doubles. Call it
// PlayerBardiasNightmare. (Bonus points for making it subtle instead of always
// double sixes. Obviously it would pull out the doubles towards the end if
// necessary, but would roll double threes instead of double sixes if it only
// needed threes.)
package ai
