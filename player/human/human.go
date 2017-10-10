// Package human contains a human-controlled tic-tac-toe player.
package human

import (
	"context"

	"github.com/shurcooL/tictactoe"
)

// NewPlayer creates a human-controlled player.
func NewPlayer() (tictactoe.Player, error) {
	return player{chosenMove: make(chan tictactoe.Move)}, nil
}

type player struct {
	chosenMove chan tictactoe.Move
}

// Name of player.
func (player) Name() string {
	return "Human Player"
}

// Play takes a tic-tac-toe board b and returns the next move
// for this player. Its mark is either X or O.
// ctx is expected to have a deadline set, and Play may take time
// to "think" until deadline is reached before returning.
func (p player) Play(ctx context.Context, b tictactoe.Board, mark tictactoe.State) (tictactoe.Move, error) {
	// Outsource our decision-making process to the human.
	// They know what they're doing. Hopefully.
	return <-p.chosenMove, nil
}

func (p player) CellClick(index int) {
	p.chosenMove <- tictactoe.Move(index)
}
