// Scalable Vector Graphics (SVG) visualizations of backgammon
package svg

import (
	"fmt"
	"sort"

	"github.com/chandler37/gobackgammon/brd"
)

type Drawer interface {
	Start(w, h int)
	End()
	Rect(x int, y int, w int, h int, s ...string)
	CenterRect(x int, y int, w int, h int, s ...string)
	Circle(x int, y int, r int, s ...string)
	Line(x1 int, y1 int, x2 int, y2 int, s ...string)
	Polyline(x []int, y []int, s ...string)
	Text(x int, y int, t string, s ...string)
}

// Visualizes board as an SVG image of width `width` pixels.
//
// The text may be an awkward size at untested values of `width`. 240 was
// tested.
//
// Remember to first call httpResponseWriter.Header().Set("Content-Type", "image/svg+xml").
//
// Does not mutate board.
func Board(width int, board *brd.Board, drawer Drawer) {
	if width < 1 {
		width = 240
	}
	c := canvas{width, width}
	drawer.Start(c.Width, c.Height)
	defer drawer.End()
	drawer.Rect(0, 0, c.Width, c.Height, fmt.Sprintf("fill:%s", backgroundColor))
	c.makeBorder(drawer)
	for i := 12; i > 0; i-- {
		c.makeTriangle(bottom, i, &board.Pips[i], drawer)
	}
	for i := 13; i < 25; i++ {
		c.makeTriangle(top, i, &board.Pips[i], drawer)
	}
	c.makeBorneOff(top, &board.Pips[brd.BorneOffWhitePip], drawer)
	c.makeBorneOff(bottom, &board.Pips[brd.BorneOffRedPip], drawer)
	c.makeBarBackground(drawer)
	c.makeBar(top, &board.Pips[brd.BarRedPip], drawer)
	c.makeBar(bottom, &board.Pips[brd.BarWhitePip], drawer)
	c.makeDoublingCubeAndStakes(board, drawer)
	c.makeDice(board, drawer)
}

type topOrBottom bool

const (
	top    topOrBottom = true
	bottom topOrBottom = !top
)

const ( // see http://www.december.com/html/spec/colorsvghex.html
	colorForLightTriangle = "palegoldenrod"
	colorForDarkTriangle  = "indigo"
	backgroundColor       = "aqua" // cadetblue
)

type canvas struct {
	Width  int // divided into 18 equal pieces
	Height int
}

const borderThickness = 2

func (c canvas) innerWidth() int {
	return c.Width - 2*borderThickness
}

func (c canvas) innerHeight() int {
	return c.Height - 2*borderThickness
}

var pointNumberToColumn = map[int]int{
	1:  15,
	2:  14,
	3:  13,
	4:  12,
	5:  11,
	6:  10,
	7:  7,
	8:  6,
	9:  5,
	10: 4,
	11: 3,
	12: 2,
	13: 2,
	14: 3,
	15: 4,
	16: 5,
	17: 6,
	18: 7,
	19: 10,
	20: 11,
	21: 12,
	22: 13,
	23: 14,
	24: 15,
}

func (c canvas) column(i int) int {
	if i < 0 || i > 18 {
		panic(fmt.Sprintf("bad args %d", i))
	}
	return borderThickness + i*c.innerWidth()/18
}

func (c canvas) triangleHeight() int {
	return int(float64(c.innerHeight()) / 2.5)
}

var checkerStyle = map[brd.Checker]string{
	brd.Red:   "fill:red",
	brd.White: "fill:white",
}

func (c canvas) checkerRadius() int {
	return (c.column(1) - c.column(0) - 2) / 2
}

func (c canvas) makeTriangle(where topOrBottom, ptNum int, pt *brd.Point, drawer Drawer) {
	color := colorForDarkTriangle
	if ptNum%2 == 0 {
		color = colorForLightTriangle
	}
	col := pointNumberToColumn[ptNum]
	if col < 2 || col > 15 {
		panic("bad col")
	}
	style := fmt.Sprintf("fill:%s; stroke-width:2;stroke:%s", color, color)
	xcoords := []int{c.column(col), (c.column(col) + c.column(col+1)) / 2, c.column(col+1) - 1}
	if where == top {
		drawer.Polyline(
			xcoords,
			[]int{borderThickness, borderThickness + c.triangleHeight(), borderThickness},
			style)
		for color, _ := range checkerStyle {
			for checkerNum := 0; checkerNum < pt.Num(color); checkerNum++ {
				y := borderThickness + c.checkerRadius() + checkerNum*c.triangleHeight()/7
				drawer.Circle(xcoords[1], y, c.checkerRadius(), checkerStyle[color])
			}
		}
	} else {
		drawer.Polyline(
			xcoords,
			[]int{c.Height - 1 - borderThickness, c.Height - 1 - borderThickness - c.triangleHeight(), c.Height - 1 - borderThickness},
			style)
		for color, _ := range checkerStyle {
			for checkerNum := 0; checkerNum < pt.Num(color); checkerNum++ {
				y := c.Height - borderThickness - c.checkerRadius() - checkerNum*c.triangleHeight()/7
				drawer.Circle(xcoords[1], y, c.checkerRadius(), checkerStyle[color])
			}
		}
	}
}

func (c canvas) makeBorder(drawer Drawer) {
	drawer.Line(0, 0, 0, c.Height-1, fmt.Sprintf("stroke-width:%d;stroke:black", borderThickness))                  // left
	drawer.Line(0, 0, c.Width-1, 0, fmt.Sprintf("stroke-width:%d;stroke:black", borderThickness))                   // top
	drawer.Line(c.Width-1, 0, c.Width-1, c.Height-1, fmt.Sprintf("stroke-width:%d;stroke:black", borderThickness))  // right
	drawer.Line(0, c.Height-1, c.Width-1, c.Height-1, fmt.Sprintf("stroke-width:%d;stroke:black", borderThickness)) // bottom
}

func (c canvas) makeBorneOff(where topOrBottom, pt *brd.Point, drawer Drawer) {
	// TODO(chandler37): Beautify. Right now you get overlap. If there are more
	// than a few checkers, write the number instead of showing them all.
	x := (c.column(17) + c.column(18)) / 2
	if where == top {
		for color, _ := range checkerStyle {
			for checkerNum := 0; checkerNum < pt.Num(color); checkerNum++ {
				y := borderThickness + c.checkerRadius() + checkerNum*c.triangleHeight()/7
				drawer.Circle(x, y, c.checkerRadius(), checkerStyle[color])
			}
		}
	} else {
		for color, _ := range checkerStyle {
			for checkerNum := 0; checkerNum < pt.Num(color); checkerNum++ {
				y := c.Height - borderThickness - c.checkerRadius() - checkerNum*c.triangleHeight()/7
				drawer.Circle(x, y, c.checkerRadius(), checkerStyle[color])
			}
		}
	}
}

func (c canvas) makeBar(where topOrBottom, pt *brd.Point, drawer Drawer) {
	x := (c.column(8) + c.column(10)) / 2
	if where == top {
		for color, _ := range checkerStyle {
			for checkerNum := 0; checkerNum < pt.Num(color); checkerNum++ {
				y := borderThickness + c.checkerRadius() + checkerNum*c.triangleHeight()/7
				drawer.Circle(x, y, c.checkerRadius(), checkerStyle[color])
			}
		}
	} else {
		for color, _ := range checkerStyle {
			for checkerNum := 0; checkerNum < pt.Num(color); checkerNum++ {
				y := c.Height - borderThickness - c.checkerRadius() - checkerNum*c.triangleHeight()/7
				drawer.Circle(x, y, c.checkerRadius(), checkerStyle[color])
			}
		}
	}
}

const barStyle = "fill:chartreuse"

func (c canvas) makeBarBackground(drawer Drawer) {
	drawer.Rect(c.column(8)+(c.column(8)-c.column(7))/2, borderThickness, c.column(8)-c.column(7), c.innerHeight(), barStyle)
}

// TODO(chandler37): Draw circles for the spots on the die, draw a square, and
// make the die the color red or white depending on the Roller.
func (c canvas) makeDie(i int, die brd.Die, player brd.Checker, drawer Drawer) {
	x := borderThickness + c.innerWidth()/3
	if player == brd.Red {
		x = borderThickness + 2*c.innerWidth()/3
	}
	dieWidth := c.innerWidth() / 60
	if die < 1 || die > 6 {
		panic(fmt.Sprintf("bad roll %d", die))
	}
	fontSize := 36
	if c.Width <= 240 {
		fontSize = 20
	}
	drawer.Text(
		x+dieWidth+3*i*dieWidth,
		c.Height/2,
		fmt.Sprintf("%d", int(die)),
		fmt.Sprintf("text-anchor:middle;font-size:%dpx;%s", fontSize, checkerStyle[player]))
}

func (c canvas) makeDice(board *brd.Board, drawer Drawer) {
	dice := make([]brd.Die, 0, 4)
	for _, die := range board.Roll {
		if die > 0 {
			dice = append(dice, die)
		}
	}
	for _, die := range board.RollUsed {
		if die > 0 {
			dice = append(dice, die)
		}
	}
	if dice[0] == dice[1] {
		dice = []brd.Die{dice[0], dice[0]}
	}
	sort.Slice(
		dice,
		func(i, j int) bool {
			return dice[i] > dice[j]
		})
	for i, die := range dice {
		c.makeDie(i, die, board.Roller, drawer)
	}
}

// TODO(chandler37): font-size should depend on c.innerHeight()

// TODO(chandler37): draw some borders to separate the borne off and doubling
// cube areas from the main board.

func (c canvas) makeDoublingCubeAndStakes(board *brd.Board, drawer Drawer) {
	if board.WhiteCanDouble || board.RedCanDouble {
		w := c.column(1) - c.column(0)
		x := (c.column(0) + c.column(1)) / 2
		y := c.Height / 2
		drawer.CenterRect(x, y, w, w, "fill:white")
		fontSize := 24
		if c.Width <= 240 {
			fontSize = 16
		}
		drawer.Text(
			x, y+c.Height/44, fmt.Sprintf("%d", board.Stakes),
			fmt.Sprintf("text-anchor:middle;font-size:%dpx;fill:black", fontSize))
	}
	goal := ""
	if g := board.MatchScore.Goal; g > 0 {
		goal = fmt.Sprintf(" Goal: %d", g)
	}
	if board.MatchScore.WhiteScore > 0 || board.MatchScore.RedScore > 0 {
		// TODO(chandler37): Beautify.
		drawer.Text(
			c.column(2), c.Height/2-c.Height/16,
			fmt.Sprintf(
				"Match Score White:%d Red:%d%s",
				board.MatchScore.WhiteScore,
				board.MatchScore.RedScore,
				goal),
			"font-size:16px;fill:black")
	}
}
