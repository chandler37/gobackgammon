package json

import (
	"math/rand"
	"testing"
	"time"

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
		example{
			373737,
			func(b *brd.Board) {
				b.Stakes = 4
			},
			`{"r":"41","st":2,"wd":1,"rd":1,"p":"r","p1":"W2","p6":"r5","p8":"r3","p12":"W5","p13":"r5","p17":"W3","p19":"W5","p24":"r2"}`,
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

func setUpTrickyBoardThatIsNotOurDictionary(b *brd.Board) {
	b.MatchScore.Goal = 6
	b.MatchScore.RedScore = 3
	b.Pips = brd.Points28{}
	for i := 1; i < 14; i++ {
		b.Pips[i].Reset(1, brd.Red)
	}
	for i := 17; i < 25; i++ {
		b.Pips[i].Reset(1, brd.White)
	}
	b.Pips[brd.BorneOffRedPip].Reset(2, brd.Red)
	b.Pips[brd.BorneOffWhitePip].Reset(5, brd.White)
	b.Pips[brd.BarWhitePip].Reset(2, brd.White)
}

func TestCompression(t *testing.T) {
	type example struct {
		Initializer func(*brd.Board)
		Token       string
		Length      int
	}
	examples := [...]example{
		example{
			func(b *brd.Board) {
			},
			"ggaVCe6AAvvBCOb5IlOYp4uMER4zhfsMIg_2Wrgx3EcQBRCnGinVAgIAAP__",
			60},
		example{
			func(b *brd.Board) {
				b.MatchScore.Goal = 5
				b.MatchScore.RedScore = 4
				b.MatchScore.NoCrawfordRule = true
				b.Pips = brd.Points28{}
				for i := 1; i < 14; i++ {
					b.Pips[i].Reset(1, brd.White)
				}
				for i := 17; i < 25; i++ {
					b.Pips[i].Reset(1, brd.Red)
				}
				b.Pips[brd.BorneOffRedPip].Reset(1, brd.Red)
				b.Pips[brd.BorneOffWhitePip].Reset(1, brd.White)
				b.Pips[brd.BarRedPip].Reset(6, brd.Red)
				b.Pips[brd.BarWhitePip].Reset(1, brd.White)
			},
			"ggaVyWhAEQgoQAAAAP__",
			20}, // tiny because we based our dictionary on this
		example{
			setUpTrickyBoardThatIsNotOurDictionary,
			"hNBLDYAwAARRT-VjZX3Mvd5JCDvlVgvvozo3UFelqBSVolJUikpRKSpFpVAKpVAKpVAqSkWpKBWlolSUilJZUmNRjR_Q_QIdcz4BAAD__w",
			106},
	}

	for _, ex := range examples {
		rand.Seed(37)
		b := brd.New(true)
		ex.Initializer(b)
		if iv := b.Invalidity(brd.IgnoreRollValidity); iv != "" {
			t.Fatalf("invalid board %v", iv)
		}
		ser, err := Serialize(b)
		if err != nil {
			t.Fatalf("err=%v", err)
		}
		ct := Compress(ser)
		if ct != ex.Token {
			t.Errorf("compresstoken=%v", ct)
		}
		if len(ct) != ex.Length {
			t.Errorf("len(compresstoken)=%v but ex.Length=%v", len(ct), ex.Length)
		}
		dct, err := Decompress(ct)
		if err != nil {
			t.Fatalf("err=%v", err)
		}
		if dct != ser {
			t.Errorf("dct=%v not %v", dct, ser)
		}
	}
}

// TODO(chandler37): As a teaching tool, enumerate a list of Boards where there
// are two nearly equally good moves but one involves contact.  Similarly,
// enumerate boards where doubling is just barely warranted.

// TODO(chandler37): Support https://en.wikipedia.org/wiki/Hypergammon

// TODO(chandler37): Support Jacoby rule

// TODO(chandler37): Support table stakes

// TODO(chandler37): Rename continuations plies. Talk about lookahead in terms of 0-ply, 1-ply, etc.

func BenchmarkCompression3(b *testing.B) {
	rand.Seed(time.Now().UnixNano())
	board := brd.New(false)
	setUpTrickyBoardThatIsNotOurDictionary(board)
	ser, err := Serialize(board)
	if err != nil {
		b.Fatalf("err=%v", err)
	}
	for n := 0; n < b.N; n++ {
		_, err := Decompress(helpCompress(ser, 3))
		if err != nil {
			b.Fatalf("err=%v", err)
		}
	}
}

func BenchmarkCompression9(b *testing.B) {
	rand.Seed(time.Now().UnixNano())
	board := brd.New(false)
	setUpTrickyBoardThatIsNotOurDictionary(board)
	ser, err := Serialize(board)
	if err != nil {
		b.Fatalf("err=%v", err)
	}
	for n := 0; n < b.N; n++ {
		_, err := Decompress(Compress(ser))
		if err != nil {
			b.Fatalf("err=%v", err)
		}
	}
}
