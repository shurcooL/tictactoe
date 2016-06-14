package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/shurcooL/htmlg"
	ttt "github.com/shurcooL/tictactoe"
	"golang.org/x/net/context"
	"honnef.co/go/js/dom"
)

// playGame plays a game of tic-tac-toe with 2 players until the end (Condition != ttt.NotEnd),
// or until an error happens. players[0] always goes first.
func playGame(players [2]player) (ttt.Condition, error) {
	var board ttt.Board // Start with an empty board.

	fmt.Println()
	fmt.Println(board)
	if runtime.GOARCH == "js" {
		var document = dom.GetWindow().Document().(dom.HTMLDocument)
		document.Body().SetInnerHTML(string(htmlg.Render(page{board: board, players: players}.Render()...)))
	}

	for i := 0; ; i++ {
		err := playerTurn(&board, players[i%2])
		if err != nil {
			if runtime.GOARCH == "js" {
				var document = dom.GetWindow().Document().(dom.HTMLDocument)
				document.Body().SetInnerHTML(string(htmlg.Render(page{board: board, errorMessage: err.Error(), players: players}.Render()...)))
			}
			return 0, err
		}
		condition := board.Condition()

		fmt.Println()
		fmt.Println(board)
		if runtime.GOARCH == "js" {
			var document = dom.GetWindow().Document().(dom.HTMLDocument)
			document.Body().SetInnerHTML(string(htmlg.Render(page{board: board, condition: condition, players: players}.Render()...)))
		}

		if condition != ttt.NotEnd {
			return condition, nil
		}
	}
}

// playerTurn gets the player p's move and applies it to board b.
func playerTurn(b *ttt.Board, player player) error {
	const timePerTurn = 3 * time.Second

	move, err := playerMove(*b, player, timePerTurn)
	if err != nil {
		return fmt.Errorf("player %v (%s) failed to make a move: %v", player.Mark, player.Name(), err)
	}
	if err := move.Validate(); err != nil {
		return fmt.Errorf("player %v (%s) made a move that isn't valid: %v", player.Mark, player.Name(), err)
	}

	err = b.Apply(move, player.Mark)
	if err != nil {
		return fmt.Errorf("player %v (%s) made a move that isn't legal: %v", player.Mark, player.Name(), err)
	}
	return nil
}

// playerMove gets the player p's move, enforcing the timeout.
func playerMove(b ttt.Board, p player, timeout time.Duration) (ttt.Move, error) {
	type moveError struct {
		ttt.Move
		err error
	}
	resultCh := make(chan moveError, 1)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// We can't trust the player not to misbehave and just ignore the timeout, causing
	// the game to stall. So we let it play inside a goroutine, and monitor ctx.Done()
	// channel ourselves. No one wants a slowpoke to hold the game up! :)
	go func() {
		move, err := p.Play(ctx, b)
		resultCh <- moveError{move, err}
	}()

	var result moveError
	select {
	case result = <-resultCh:
		return result.Move, result.err
	case <-ctx.Done():
		return 0, fmt.Errorf("took more than allotted time of %v", timeout)
	}
}
