package json

import (
	"math/rand"
	"testing"

	"github.com/chandler37/gobackgammon/brd"
)

func TestMath(t *testing.T) {
	type example struct {
		N      int
		Result uint64
	}
	examples := [...]example{
		example{0, 1},
		example{1, 2},
		example{2, 4},
		example{10, 1024},
		example{20, 1024 * 1024},
	}
	for _, ex := range examples {
		if a := twoToThePower(ex.N); a != ex.Result {
			t.Errorf("2**N: %v %v", ex, a)
		}
		if a, err := log2(ex.Result); err != nil || a != ex.N {
			t.Errorf("log2: %v %v", ex, a)
		}
	}
}

func TestSer(t *testing.T) {
	type example struct {
		Seed        int64
		Initializer func(*brd.Board)
		Json        string
	}
	examples := [...]example{
		example{
			37,
			func(b *brd.Board) {
			},
			`{"r":"64","wd":1,"rd":1,"p":"W","p1":"W2","p6":"r5","p8":"r3","p12":"W5","p13":"r5","p17":"W3","p19":"W5","p24":"r2"}`,
		},
		example{
			3737,
			func(b *brd.Board) {
				b.MatchScore.Goal = 1
			},
			`{"r":"54","wd":1,"rd":1,"p":"W","p1":"W2","p6":"r5","p8":"r3","p12":"W5","p13":"r5","p17":"W3","p19":"W5","p24":"r2","s":{"g":1}}`,
		},
		example{
			373737,
			func(b *brd.Board) {
				b.WhiteCanDouble = false
				b.RedCanDouble = false
				b.Pips[brd.BarRedPip].Reset(1, brd.Red)
				b.Pips[brd.BorneOffRedPip].Reset(1, brd.Red)
				b.Pips[6].Subtract()
				b.Pips[6].Subtract()
				b.MatchScore.WhiteScore = 1
				b.MatchScore.AlreadyPlayedCrawfordGame = true
				b.MatchScore.NoCrawfordRule = true
			},
			`{"r":"41","p":"r","p1":"W2","p6":"r3","p8":"r3","p12":"W5","p13":"r5","p17":"W3","p19":"W5","p24":"r2","p25":"r","p27":"r","s":{"w":1,"n":1,"a":1}}`,
		},
	}
	for _, ex := range examples {
		rand.Seed(ex.Seed)
		b := brd.New(true)
		ex.Initializer(b)
		if iv := b.Invalidity(brd.IgnoreRollValidity); iv != "" {
			t.Errorf("invalidity for %v is %v", ex, iv)
		}
		if s, err := Serialize(b); err != nil || s != ex.Json {
			t.Errorf("\nex=%v\nerr=%v\n%v\nserialized to %v", ex, err, b, s)
		} else {
			if bb, err := Deserialize(s); err != nil || !bb.Equals(*b) || !b.Equals(*bb) {
				t.Errorf("ex=%v\nbb=%v\nerr=%v\nb=%v", ex, bb, err, b)
			}
		}
	}
}
