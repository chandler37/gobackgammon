package svg

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/chandler37/gobackgammon/brd"
)

type testDrawer struct {
	Actions []string
}

func (d *testDrawer) Start(w, h int) {
	d.Actions = append(d.Actions, "Start")
}

func (d *testDrawer) End() {
	d.Actions = append(d.Actions, "End")
}

func (d *testDrawer) Rect(x int, y int, w int, h int, s ...string) {
	d.Actions = append(d.Actions, "Rect")
}

func (d *testDrawer) CenterRect(x int, y int, w int, h int, s ...string) {
	d.Actions = append(d.Actions, "CenterRect")
}

func (d *testDrawer) Circle(x int, y int, r int, s ...string) {
	d.Actions = append(d.Actions, "Circle")
}

func (d *testDrawer) Line(x1 int, y1 int, x2 int, y2 int, s ...string) {
	d.Actions = append(d.Actions, "Line")
}

func (d *testDrawer) Polyline(x []int, y []int, s ...string) {
	d.Actions = append(d.Actions, "Polyline")
}

func (d *testDrawer) Text(x int, y int, t string, s ...string) {
	d.Actions = append(d.Actions, "Text")
}

var gold = `Start
Rect
Line
Line
Line
Line
Polyline
Circle
Circle
Circle
Circle
Circle
Polyline
Polyline
Polyline
Polyline
Circle
Circle
Circle
Polyline
Polyline
Circle
Circle
Circle
Circle
Circle
Polyline
Polyline
Polyline
Polyline
Polyline
Circle
Circle
Polyline
Circle
Circle
Circle
Circle
Circle
Polyline
Polyline
Polyline
Polyline
Circle
Circle
Circle
Polyline
Polyline
Circle
Circle
Circle
Circle
Circle
Polyline
Polyline
Polyline
Polyline
Polyline
Circle
Circle
Rect
CenterRect
Text
Text
Text
End`

func TestBoard(t *testing.T) {
	rand.Seed(37)
	b := brd.New(true)
	drawer := testDrawer{[]string{}}
	Board(240, b, &drawer)
	if x := strings.Join(drawer.Actions, "\n"); x != gold {
		t.Errorf("x=%v", x)
	}
}
