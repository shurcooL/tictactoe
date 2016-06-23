// Package bad contains a bad tic-tac-toe player.
package bad

import (
	"fmt"
	"time"

	"github.com/shurcooL/tictactoe"
	"golang.org/x/net/context"
)

// NewPlayer creates a bad player.
func NewPlayer() (tictactoe.Player, error) {
	return player{}, nil
}

type player struct{}

// Name of player.
func (player) Name() string {
	return "Bad Player"
}

// Play takes a tic-tac-toe board b and returns the next move
// for this player. Its mark is either X or O.
// ctx is expected to have a deadline set, and Play may take time
// to "think" until deadline is reached before returning.
func (player) Play(ctx context.Context, b tictactoe.Board, mark tictactoe.State) (tictactoe.Move, error) {
	// Who cares about some deadline?
	_, _ = ctx.Deadline()

	// Take lots of time to think about what to do...
	time.Sleep(20 * time.Second)

	return 0, fmt.Errorf("... I have no idea how to play tic-tac-toe. :( Can you help me?")
}
