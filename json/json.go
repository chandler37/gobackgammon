// JSON serialization and deserialization of backgammon
package json

import (
	"bytes"
	"compress/flate"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/chandler37/gobackgammon/brd"
)

func Serialize(b *brd.Board) (string, error) {
	cb := &compactBoard{}
	cb.serializeMatchScore(&b.MatchScore)
	if b.WhiteCanDouble {
		cb.WhiteCanDouble = 1
	}
	if b.RedCanDouble {
		cb.RedCanDouble = 1
	}
	cb.Roller = "W"
	if b.Roller == brd.Red {
		cb.Roller = "r"
	}
	cb.RollUsed = makeCompactRoll(&b.RollUsed)
	cb.Roll = makeCompactRoll(&b.Roll)
	var err error
	cb.StakesLog2, err = log2(uint64(b.Stakes))
	if err != nil {
		return "", fmt.Errorf("Bad Stakes: %v", err)
	}
	if len(b.Pips) != 28 {
		panic("brd changed")
	}
	i := 0
	cb.P0 = makeCompactPoint(&b.Pips[i])
	i++
	cb.P1 = makeCompactPoint(&b.Pips[i])
	i++
	cb.P2 = makeCompactPoint(&b.Pips[i])
	i++
	cb.P3 = makeCompactPoint(&b.Pips[i])
	i++
	cb.P4 = makeCompactPoint(&b.Pips[i])
	i++
	cb.P5 = makeCompactPoint(&b.Pips[i])
	i++
	cb.P6 = makeCompactPoint(&b.Pips[i])
	i++
	cb.P7 = makeCompactPoint(&b.Pips[i])
	i++
	cb.P8 = makeCompactPoint(&b.Pips[i])
	i++
	cb.P9 = makeCompactPoint(&b.Pips[i])
	i++
	cb.P10 = makeCompactPoint(&b.Pips[i])
	i++
	cb.P11 = makeCompactPoint(&b.Pips[i])
	i++
	cb.P12 = makeCompactPoint(&b.Pips[i])
	i++
	cb.P13 = makeCompactPoint(&b.Pips[i])
	i++
	cb.P14 = makeCompactPoint(&b.Pips[i])
	i++
	cb.P15 = makeCompactPoint(&b.Pips[i])
	i++
	cb.P16 = makeCompactPoint(&b.Pips[i])
	i++
	cb.P17 = makeCompactPoint(&b.Pips[i])
	i++
	cb.P18 = makeCompactPoint(&b.Pips[i])
	i++
	cb.P19 = makeCompactPoint(&b.Pips[i])
	i++
	cb.P20 = makeCompactPoint(&b.Pips[i])
	i++
	cb.P21 = makeCompactPoint(&b.Pips[i])
	i++
	cb.P22 = makeCompactPoint(&b.Pips[i])
	i++
	cb.P23 = makeCompactPoint(&b.Pips[i])
	i++
	cb.P24 = makeCompactPoint(&b.Pips[i])
	i++
	cb.P25 = makeCompactPoint(&b.Pips[i])
	i++
	cb.P26 = makeCompactPoint(&b.Pips[i])
	i++
	cb.P27 = makeCompactPoint(&b.Pips[i])
	i++

	x, err := json.Marshal(cb)
	if err != nil {
		panic(err)
	}
	return string(x), nil
}

func Deserialize(s string) (*brd.Board, error) {
	cb := compactBoard{}
	err := json.Unmarshal([]byte(s), &cb)
	if err != nil {
		return nil, err
	}
	b := brd.New(false)
	cbPoints := [...]string{
		cb.P0, cb.P1, cb.P2, cb.P3, cb.P4, cb.P5, cb.P6, cb.P7, cb.P8, cb.P9,
		cb.P10, cb.P11, cb.P12, cb.P13, cb.P14, cb.P15, cb.P16, cb.P17, cb.P18, cb.P19,
		cb.P20, cb.P21, cb.P22, cb.P23, cb.P24, cb.P25, cb.P26, cb.P27}
	for i, pt := range cbPoints {
		k, color, err := parseCompactPoint(pt)
		if err != nil {
			return nil, err
		}
		b.Pips[i].Reset(k, color)
	}
	switch cb.Roller {
	case "W":
		b.Roller = brd.White
	case "r":
		b.Roller = brd.Red
	default:
		return nil, fmt.Errorf("bad Roller in %v", s)
	}
	b.RollUsed, err = parseRoll(cb.RollUsed)
	if err != nil {
		return nil, fmt.Errorf("bad RollUsed in %v: %v", s, err)
	}
	b.Roll, err = parseRoll(cb.Roll)
	if err != nil {
		return nil, fmt.Errorf("bad Roll in %v: %v", s, err)
	}
	b.Stakes = int(twoToThePower(cb.StakesLog2))
	if cb.WhiteCanDouble != 0 {
		b.WhiteCanDouble = true
	} else {
		b.WhiteCanDouble = false
	}
	if cb.RedCanDouble != 0 {
		b.RedCanDouble = true
	} else {
		b.RedCanDouble = false
	}
	cb.deserializeMatchScore(&b.MatchScore)
	if iv := b.Invalidity(brd.IgnoreRollValidity); iv != "" {
		return nil, fmt.Errorf("invalid board: %v", iv)
	}
	return b, nil
}

// Given the result of Serialize(), returns a URL-safe base64 representation of
// a flate-compressed form of the result (using a custom flate
// dictionary). (See the `compress/flate` library.)
func Compress(boardSerialization string) string {
	return helpCompress(boardSerialization, flate.BestCompression)
}

func Decompress(compressedUrlSafeBase64Board string) (string, error) {
	rawToken, err := base64.RawURLEncoding.DecodeString(compressedUrlSafeBase64Board)
	if err != nil {
		return "", fmt.Errorf("Bad base64: %v", err)
	}
	zr := flate.NewReaderDict(bytes.NewReader(rawToken), compressionDict)
	defer zr.Close()
	someBytes, err := ioutil.ReadAll(zr)
	if err != nil {
		return "", err
	}
	return string(someBytes), nil
}

var compressionDict = []byte(`{"r":"6666","wd":1,"rd":1,"p":"W","p0":"W","p1":"W","p2":"W","p3":"W","p4":"W","p5":"W","p6":"W","p7":"W","p8":"W","p9":"W","p10":"W","p11":"W","p12":"W","p13":"W","p17":"r","p18":"r","p19":"r","p20":"r","p21":"r","p22":"r","p23":"r","p24":"r","p25":"r","p26":"W","p27":"r6","s":{"g":5,"r":4,"n":1}}`)

// We could also try `{"r":"41","p":"r","p1":"W2","p6":"r3","p8":"r3","p12":"W5","p13":"r5","p17":"W3","p19":"W5","p24":"r2","p25":"r","p27":"r","s":{"w":1,"n":1,"a":1}}`

func helpCompress(serialization string, compressionLevel int) string {
	var b bytes.Buffer
	compressor, err := flate.NewWriterDict(&b, compressionLevel, compressionDict)
	if err != nil {
		panic(err) // only happens with invalid compression level
	}
	_, err = compressor.Write([]byte(serialization))
	if err != nil {
		panic(err)
	}
	compressor.Close()
	return base64.RawURLEncoding.EncodeToString(b.Bytes())
}

// Example: {"ru":"3","ro":"1","st":6,"wd":0,"rd":1,"p":"W","p0":"W","p1":"W15","s":{"g":3,"w":1,"r":0,"c":0,"a":0}}
type compactBoard struct {
	RollUsed       string        `json:"ru,omitempty"`
	Roll           string        `json:"r,omitempty"`
	StakesLog2     int           `json:"st,omitempty"`
	WhiteCanDouble int           `json:"wd,omitempty"`
	RedCanDouble   int           `json:"rd,omitempty"`
	Roller         string        `json:"p,omitempty"`
	P0             string        `json:"p0,omitempty"`
	P1             string        `json:"p1,omitempty"`
	P2             string        `json:"p2,omitempty"`
	P3             string        `json:"p3,omitempty"`
	P4             string        `json:"p4,omitempty"`
	P5             string        `json:"p5,omitempty"`
	P6             string        `json:"p6,omitempty"`
	P7             string        `json:"p7,omitempty"`
	P8             string        `json:"p8,omitempty"`
	P9             string        `json:"p9,omitempty"`
	P10            string        `json:"p10,omitempty"`
	P11            string        `json:"p11,omitempty"`
	P12            string        `json:"p12,omitempty"`
	P13            string        `json:"p13,omitempty"`
	P14            string        `json:"p14,omitempty"`
	P15            string        `json:"p15,omitempty"`
	P16            string        `json:"p16,omitempty"`
	P17            string        `json:"p17,omitempty"`
	P18            string        `json:"p18,omitempty"`
	P19            string        `json:"p19,omitempty"`
	P20            string        `json:"p20,omitempty"`
	P21            string        `json:"p21,omitempty"`
	P22            string        `json:"p22,omitempty"`
	P23            string        `json:"p23,omitempty"`
	P24            string        `json:"p24,omitempty"`
	P25            string        `json:"p25,omitempty"`
	P26            string        `json:"p26,omitempty"`
	P27            string        `json:"p27,omitempty"`
	MatchScore     *compactScore `json:"s,omitempty"`
}

type compactScore struct {
	Goal                      int `json:"g,omitempty"`
	WhiteScore                int `json:"w,omitempty"`
	RedScore                  int `json:"r,omitempty"`
	NoCrawfordRule            int `json:"n,omitempty"`
	AlreadyPlayedCrawfordGame int `json:"a,omitempty"`
}

// "W15" for fifteen White or "r" for one Red or "" for an empty Point
func makeCompactPoint(pt *brd.Point) string {
	for i := len(*pt) - 1; i >= 0; i-- {
		if pt[i] == brd.NoChecker {
			continue
		}
		c := "r"
		if pt[i] == brd.White {
			c = "W"
		}
		if i == 0 {
			return c
		} else {
			return fmt.Sprintf("%s%d", c, i+1)
		}
	}
	return ""
}

func makeCompactRoll(r *brd.Roll) (result string) {
	for _, die := range *r {
		if die != brd.ZeroDie {
			result = result + fmt.Sprintf("%d", die)
		}
	}
	return result
}

func twoToThePower(i int) (result uint64) {
	result = 1
	for j := 0; j < i; j++ {
		result *= 2
	}
	return
}

func log2(i uint64) (int, error) {
	switch i {
	case 1:
		return 0, nil
	case 2:
		return 1, nil
	case 4:
		return 2, nil
	case 8:
		return 3, nil
	case 16:
		return 4, nil
	case 32:
		return 5, nil
	case 64:
		return 6, nil
	case 128:
		return 7, nil
	case 256:
		return 8, nil
	case 512:
		return 9, nil
	case 1024:
		return 10, nil
	case 2048:
		return 11, nil
	case 4096:
		return 12, nil
	case 8192:
		return 13, nil
	case 16384:
		return 14, nil
	case 32768:
		return 15, nil
	case 65536:
		return 16, nil
	case 131072:
		return 17, nil
	case 262144:
		return 18, nil
	case 524288:
		return 19, nil
	case 1048576:
		return 20, nil
	case 2097152:
		return 21, nil
	case 4194304:
		return 22, nil
	case 8388608:
		return 23, nil
	case 16777216:
		return 24, nil
	case 33554432:
		return 25, nil
	case 67108864:
		return 26, nil
	case 134217728:
		return 27, nil
	case 268435456:
		return 28, nil
	case 536870912:
		return 29, nil
	default:
		return 0, fmt.Errorf("log2: illegal input %d. This program is not equipped for your level of tomfoolery.", i)
	}
}

func (cb *compactBoard) serializeMatchScore(s *brd.Score) {
	cs := compactScore{}
	cs.Goal = s.Goal
	cs.WhiteScore = s.WhiteScore
	cs.RedScore = s.RedScore
	if s.NoCrawfordRule {
		cs.NoCrawfordRule = 1
	}
	if s.AlreadyPlayedCrawfordGame {
		cs.AlreadyPlayedCrawfordGame = 1
	}
	zero := compactScore{}
	if cs != zero {
		cb.MatchScore = &cs
	}
}

func parseCompactPoint(s string) (int, brd.Checker, error) {
	if len(s) < 1 {
		return 0, brd.White, nil // arbitrarily
	}
	color := brd.White
	if s[0] == 'W' {
	} else if s[0] == 'r' {
		color = brd.Red
	} else {
		return 0, brd.NoChecker, fmt.Errorf("bad point %v because of color", s)
	}
	if len(s) == 1 {
		return 1, color, nil
	}
	k, err := strconv.Atoi(s[1:len(s)])
	if err != nil || k < 1 {
		return 0, brd.NoChecker, fmt.Errorf("bad point %v", s)
	}
	return k, color, nil
}

func parseRoll(s string) (result brd.Roll, err error) {
	i := 0
	for _, b := range s {
		if i >= 4 {
			return brd.Roll{}, fmt.Errorf("too many dice in %v", s)
		}
		switch b {
		case '1':
			result[i] = 1
			i++
		case '2':
			result[i] = 2
			i++
		case '3':
			result[i] = 3
			i++
		case '4':
			result[i] = 4
			i++
		case '5':
			result[i] = 5
			i++
		case '6':
			result[i] = 6
			i++
		default:
			return brd.Roll{}, fmt.Errorf("bad roll %v", s)
		}
	}
	if result[0] < result[1] {
		// There are up to 4 dice but there are only at most two distinct dice.
		result[0], result[1] = result[1], result[0]
	}
	return
}

func (cb *compactBoard) deserializeMatchScore(score *brd.Score) {
	if cb.MatchScore == nil {
		return
	}
	score.Goal = cb.MatchScore.Goal
	score.WhiteScore = cb.MatchScore.WhiteScore
	score.RedScore = cb.MatchScore.RedScore
	if cb.MatchScore.NoCrawfordRule != 0 {
		score.NoCrawfordRule = true
	}
	if cb.MatchScore.AlreadyPlayedCrawfordGame != 0 {
		score.AlreadyPlayedCrawfordGame = true
	}
}
