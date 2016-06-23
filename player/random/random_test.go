package random_test

import (
	"testing"
	"time"

	ttt "github.com/shurcooL/tictactoe"
	"github.com/shurcooL/tictactoe/player/random"
	"golang.org/x/net/context"
)

func Test(t *testing.T) {
	// This board has only one free cell, so there's only one legal move.
	b := ttt.Board{
		Cells: [9]ttt.State{
			ttt.X, ttt.X, ttt.O,
			ttt.O, ttt.F, ttt.X,
			ttt.O, ttt.X, ttt.O,
		},
	}
	mark := ttt.X
	want := ttt.Move(4)

	player, err := random.NewPlayer()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	move, err := player.Play(ctx, b, mark)
	cancel()
	if err != nil {
		t.Fatal(err)
	}
	if move != want {
		t.Errorf("not the expected move: %v", move)
	}
}
