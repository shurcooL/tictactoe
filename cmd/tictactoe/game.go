package main

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/gopherjs/gopherjs/js"
	"github.com/shurcooL/htmlg"
	ttt "github.com/shurcooL/tictactoe"
	"honnef.co/go/js/dom"
)

// timePerTurn is the time each player gets to think per turn.
const timePerTurn = 5 * time.Second

// playGame plays a game of tic-tac-toe with 2 players until the end (Condition != ttt.NotEnd),
// or until an error happens. players[0] always goes first.
func playGame(players [2]player) (ttt.Condition, error) {
	// When a board cell is clicked, its [0, 9) index is sent to this channel.
	var cellClick chan int

	if runtime.GOARCH == "js" {
		cellClick = make(chan int)
		js.Global.Set("CellClick", func(index int) {
			select {
			case cellClick <- index:
			default:
			}
		})
	}

	// Start with an empty board.
	var board ttt.Board
	var condition ttt.Condition

	fmt.Println()
	fmt.Println(board)

	for i := 0; condition == ttt.NotEnd; i = (i + 1) % 2 {
		if runtime.GOARCH == "js" {
			// Draw page at start of turn.
			var document = dom.GetWindow().Document().(dom.HTMLDocument)
			_, isCellClicker := players[i].Player.(ttt.CellClicker)
			document.Body().SetInnerHTML(htmlg.Render(page{Board: board, Turn: players[i].Mark, Clickable: isCellClicker, Condition: condition, Players: players}.Render()...))
			runtime.Gosched()
		}

		turnStart := time.Now()

		err := playerTurn(&board, players[i], cellClick)
		if err != nil {
			if runtime.GOARCH == "js" {
				// Draw page on error.
				var document = dom.GetWindow().Document().(dom.HTMLDocument)
				document.Body().SetInnerHTML(htmlg.Render(page{Board: board, ErrorMessage: err.Error(), Players: players}.Render()...))
			}
			return 0, err
		}

		condition = board.Condition()

		// Enforce a minimum of 1 second per turn.
		if untilTurnEnd := time.Second - time.Since(turnStart); untilTurnEnd > 0 {
			if runtime.GOARCH == "js" {
				// Draw page after player finished turn.
				var document = dom.GetWindow().Document().(dom.HTMLDocument)
				document.Body().SetInnerHTML(htmlg.Render(page{Board: board, Condition: condition, Players: players}.Render()...))
			}

			time.Sleep(untilTurnEnd)
		}

		fmt.Println()
		fmt.Println(board)
	}

	// At this point, the game is over.
	if runtime.GOARCH == "js" {
		// Draw page on end of game.
		var document = dom.GetWindow().Document().(dom.HTMLDocument)
		document.Body().SetInnerHTML(htmlg.Render(page{Board: board, Condition: condition, Players: players}.Render()...))
	}

	return condition, nil
}

// playerTurn gets the player p's move and applies it to board b.
func playerTurn(b *ttt.Board, player player, cellClick <-chan int) error {
	move, err := playerMove(*b, player, timePerTurn, cellClick)
	if err != nil {
		return fmt.Errorf("player %v (%s) failed to make a move: %v", player.Mark, player.Name(), err)
	}
	if err := move.Valid(); err != nil {
		return fmt.Errorf("player %v (%s) made a move that isn't valid: %v", player.Mark, player.Name(), err)
	}

	err = b.Apply(move, player.Mark)
	if err != nil {
		return fmt.Errorf("player %v (%s) made a move that isn't legal: %v", player.Mark, player.Name(), err)
	}

	return nil
}

// playerMove gets the player p's move, enforcing the timeout.
func playerMove(b ttt.Board, p player, timeout time.Duration, cellClick <-chan int) (ttt.Move, error) {
	type moveError struct {
		ttt.Move
		err error
	}
	resultCh := make(chan moveError, 1)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// We can't trust the player not to misbehave and just ignore the timeout, causing
	// the game to stall. So we let it play inside a goroutine, and monitor ctx.Done()
	// channel ourselves. No one wants a slowpoke to hold the game up! :) Also catch panics.
	go func() {
		defer func() {
			if e := recover(); e != nil {
				resultCh <- moveError{err: fmt.Errorf("panic: %v", e)}
			}
		}()
		move, err := p.Play(ctx, b, p.Mark)
		resultCh <- moveError{move, err}
	}()

	for {
		select {
		case result := <-resultCh:
			return result.Move, result.err
		case index := <-cellClick:
			if p, ok := p.Player.(ttt.CellClicker); ok {
				p.CellClick(index)
			}
		case <-ctx.Done():
			return 0, fmt.Errorf("took more than allotted time of %v", timeout)
		}
	}
}
