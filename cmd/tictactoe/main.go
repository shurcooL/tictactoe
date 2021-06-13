// tictactoe plays a game of tic-tac-toe with two players.
//
// It's just for fun, a learning exercise.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	ttt "github.com/shurcooL/tictactoe"
)

import (
	playerx "github.com/shurcooL/tictactoe/player/random"
	// vs
	playero "github.com/shurcooL/tictactoe/player/perfect"
)

// timePerTurn is the time each player gets to think per turn.
const timePerTurn = 5 * time.Second

func main() {
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

	simulateGame([2]player{playerX, playerO})
}

type player struct {
	ttt.Player
	Mark ttt.State // Mark is either X or O.
}

// simulateGame simulates a playthrough of a game of tic-tac-toe with 2 players
// until the end (Condition != ttt.NotEnd), or until an error happens.
// players[0] always goes first.
func simulateGame(players [2]player) {
	// Start with an empty board.
	var board ttt.Board
	var condition ttt.Condition

	// When a board cell is clicked, its [0, 9) index is sent to this channel.
	cellClick := make(chan int)

	displayGameStart(board, players, cellClick)

	for i := 0; condition == ttt.NotEnd; i = (i + 1) % 2 {
		displayTurnStart(board, players, players[i], condition)

		turnStart := time.Now()

		err := playerTurn(&board, players[i], cellClick)
		if err != nil {
			displayError(board, players, err)
			return
		}

		condition = board.Condition()

		// Enforce a minimum of 1 second per turn.
		if untilTurnEnd := time.Second - time.Since(turnStart); untilTurnEnd > 0 {
			displayTurnEnding(board, players, condition)

			time.Sleep(untilTurnEnd)
		}
	}

	// At this point, the game is over.
	displayGameEnd(board, players, condition)
}

// playerTurn gets the player p's move and applies it to board b.
func playerTurn(b *ttt.Board, player player, cellClick <-chan int) error {
	move, err := playerMove(*b, player, timePerTurn, cellClick)
	if err != nil {
		return fmt.Errorf("player %v (%s) failed to make a move: %v", player.Mark, player.Name(), err)
	}

	err = b.Apply(move, player.Mark)
	if err != nil {
		return fmt.Errorf("player %v (%s) made a move that isn't valid or isn't legal: %v", player.Mark, player.Name(), err)
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
