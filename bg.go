package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/chandler37/gobackgammon/ai"
	"github.com/chandler37/gobackgammon/brd"
)

var seedOverride = flag.Int64(
	"seed",
	-1,
	"PRNG seed for those that hate variety")

var matchGoal = flag.Int64(
	"match",
	-1,
	"In -auto mode, how many points do you want to play to? Crawford rule is enabled but TODO(chandler37): Support doubling in the AI.")

var auto = flag.Bool(
	"auto",
	false,
	"Watch the computer play against itself")

var debug = flag.Bool(
	"debug",
	false,
	"Print copious amounts of AI debugging info")

var random = flag.Bool(
	"random",
	false,
	"Use an insane dart thrower instead of a conservative player.")

var automaticallyAcceptTheOnlyChoice = flag.Bool(
	"automaticallyAcceptTheOnlyChoice",
	false,
	"When there is only one legal play, take it without prompting.")

func matchIsOver(score brd.Score) bool {
	return score.RedScore >= score.Goal || score.WhiteScore >= score.Goal
}

func playAuto() {
	score := brd.Score{}
	if *matchGoal > 0 {
		score.Goal = int(*matchGoal)
	}
	for {
		board := brd.New(false)
		board.SetScore(score)
		fmt.Printf(
			"The official backgammon rules dictate the following starting configuration\nfor a game against Red ('r') and White ('W')\nand we've randomly chosen who goes first and with which roll:\n%v\n",
			board)
		fmt.Printf("\nHere is a game where Red and White both choose a random move:\n")
		numBoards := 0
		logger := func(_ interface{}, b *brd.Board) {
			numBoards++
			fmt.Printf("%v\n", b.String())
		}
		chooser := ai.MakePlayerConservative(0, nil)
		if *random {
			chooser = ai.PlayerRandom
		}
		victor, stakes, subscore := board.PlayGame(struct{}{}, chooser, logger, nil, nil)
		fmt.Printf(
			"\n\nTHE END OF A SINGLE GAME\n%v was victorious on Board number %d, winning %d points (2 is a gammon, 3 is a backgammon), with match score %v\n",
			victor, numBoards, stakes, score)
		if matchIsOver(subscore) {
			fmt.Printf(
				"\n\nTHE END OF THE MATCH\n%v was victorious with match score %v\n",
				victor, subscore)
			return
		}
		score = subscore
	}
}

func main() {
	flag.Parse()
	seed := time.Now().UnixNano()
	if *seedOverride > -1 {
		seed = *seedOverride
	}
	fmt.Printf("rand.Seed(%v)\n", seed)
	rand.Seed(seed)
	if *auto {
		playAuto()
		return
	}
	board := brd.New(false)
	fmt.Printf(
		"The official backgammon rules dictate the following starting configuration\nfor a game against Red ('r') and White ('W')\nand we've randomly chosen who goes first and with which roll:\n%v\nYou are playing Red.\n",
		board)
	if board.Roller == brd.Red {
		fmt.Printf("\nYour turn, Red!\n")
	} else {
		fmt.Printf("\nWhite goes first.\n")
	}
	numBoards := 0
	logger := func(_ interface{}, b *brd.Board) {
		numBoards++
		fmt.Printf("%v\n", b.String())
	}
	reader := bufio.NewReader(os.Stdin)
	if *debug {
		ai.StartDebugging()
	}
	redChooser := ai.MakePlayerConservative(0, nil)
	chooser := func(s []brd.Board) int {
		if s[0].Roller == brd.White {
			ai.StopDebugging()
			c := ai.MakePlayerConservative(0, nil)(s)
			if *debug {
				ai.StartDebugging()
			}
			return c
		}
		if len(s) == 0 {
			panic("will not happen")
		}
		if len(s) == 1 && *automaticallyAcceptTheOnlyChoice {
			return 0
		}
		conservativeChoice := redChooser(s)
		fmt.Printf("Which number do you choose from the following choices? (* is AI's choice, %d)\n", conservativeChoice)
		for i, b := range s {
			hint := " "
			if conservativeChoice == i {
				hint = "*"
			}
			fmt.Printf("%s%-3d: %v\n", hint, i, b.String())
		}
		for {
			fmt.Print("Enter choice number or a substring filter: ")
			text, _ := reader.ReadString('\n')
			text = strings.TrimSuffix(text, "\n")
			if text == "" {
				continue
			}
			if text == "*" {
				return conservativeChoice
			}
			choice, err := strconv.Atoi(text)
			if err != nil {
				for n, b := range s {
					bs := b.String()
					if strings.Contains(bs, text) {
						fmt.Printf("%-3d: %v\n", n, bs)
					}
				}
				continue
			}
			if choice < 0 || choice >= len(s) {
				fmt.Println("That number is out of bounds.")
				continue
			}
			return choice
		}
	}
	victor, stakes, score := board.PlayGame(struct{}{}, chooser, logger, nil, nil)
	fmt.Printf(
		"\n\nTHE END\n%v was victorious on Board number %d, winning %d points (2 is a gammon, 3 is a backgammon) with match score %v\n",
		victor, numBoards, stakes, score)
}
