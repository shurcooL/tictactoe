// +build !js

package main

import (
	"fmt"

	ttt "github.com/shurcooL/tictactoe"
)

func main() {
	run()
}

func displayGameStart(board ttt.Board, players [2]player, cellClick chan<- int) {
	fmt.Println("Tic-Tac-Toe")
	fmt.Println()
	fmt.Printf("%v (X) vs %v (O)\n", players[0].Name(), players[1].Name())
}

func displayTurnStart(board ttt.Board, players [2]player, active player, condition ttt.Condition) {
	fmt.Println()
	fmt.Println(board)
}

func displayTurnEnding(board ttt.Board, players [2]player, condition ttt.Condition) {}

func displayGameEnd(board ttt.Board, players [2]player, condition ttt.Condition) {
	fmt.Println()
	fmt.Println(board)
	fmt.Println()
	switch condition {
	case ttt.XWon:
		fmt.Printf("player X (%v) won!\n", players[0].Name())
	case ttt.OWon:
		fmt.Printf("player O (%v) won!\n", players[1].Name())
	case ttt.Tie:
		fmt.Println("game ended in a tie.")
	default:
		fmt.Println(condition)
	}
}

func displayError(board ttt.Board, players [2]player, err error) {
	fmt.Println(err)
}
