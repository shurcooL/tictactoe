// tictactoe plays a game of tic-tac-toe with two players.
//
// It's just for fun, a learning exercise.
package main

import (
	"fmt"
	"log"
	"runtime"

	ttt "github.com/shurcooL/tictactoe"
	"honnef.co/go/js/dom"
)

import (
	playerx "github.com/shurcooL/tictactoe/player/random"
	// vs
	playero "github.com/shurcooL/tictactoe/player/perfect"
)

func main() {
	switch runtime.GOARCH {
	default:
		run()
	case "js":
		var document = dom.GetWindow().Document().(dom.HTMLDocument)
		switch readyState := document.ReadyState(); readyState {
		case "loading":
			document.AddEventListener("DOMContentLoaded", false, func(dom.Event) {
				go run()
			})
		case "interactive", "complete":
			run()
		default:
			panic(fmt.Errorf("internal error: unexpected document.ReadyState value: %v", readyState))
		}
	}
}

func run() {
	playerX := player{Mark: ttt.X}
	playerO := player{Mark: ttt.O}

	var err error
	playerX.Player, err = playerx.NewPlayer()
	if err != nil {
		log.Fatalln(fmt.Errorf("failed to initialize player X: %v", err))
	}
	playerO.Player, err = playero.NewPlayer()
	if err != nil {
		log.Fatalln(fmt.Errorf("failed to initialize player O: %v", err))
	}

	fmt.Println("Tic-Tac-Toe")
	fmt.Println()
	fmt.Printf("%v (X) vs %v (O)\n", playerX.Name(), playerO.Name())
	if runtime.GOARCH == "js" {
		var document = dom.GetWindow().Document().(dom.HTMLDocument)
		document.SetTitle("Tic-Tac-Toe")
	}

	endCondition, err := playGame([2]player{playerX, playerO})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println()
	switch endCondition {
	case ttt.XWon:
		fmt.Printf("player X (%v) won!\n", playerX.Name())
	case ttt.OWon:
		fmt.Printf("player O (%v) won!\n", playerO.Name())
	case ttt.Tie:
		fmt.Println("game ended in a tie.")
	default:
		fmt.Println(endCondition)
	}
}

type player struct {
	ttt.Player
	Mark ttt.State // Mark is either X or O.
}
