// Package perfect implements a perfect tic-tac-toe player.
//
// It always wins if the opponent makes a suboptimal move
// that opens up an opportunity to guarantee a win.
// It never loses.
package perfect

import (
	"context"
	"fmt"
	"html/template"
	"math/rand"
	"time"

	ttt "github.com/shurcooL/tictactoe"
)

// NewPlayer creates a perfect player.
func NewPlayer() (ttt.Player, error) {
	return player{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

type player struct {
	rand *rand.Rand
}

// Name of player.
func (player) Name() string {
	return "Perfect Player"
}

func (player) Image() template.URL {
	return "https://raw.githubusercontent.com/shurcooL/tictactoe/master/player/perfect/gopher-fancy.png"
}

// Play takes a tic-tac-toe board b and returns the next move
// for this player. Its mark is either X or O.
// ctx is expected to have a deadline set, and Play may take time
// to "think" until deadline is reached before returning.
func (p player) Play(ctx context.Context, b ttt.Board, mark ttt.State) (ttt.Move, error) {
	if b.Condition() != ttt.NotEnd {
		return ttt.Move(-1), fmt.Errorf("board has a finished game")
	}

	stopThinking := time.Now().Add(2 * time.Second)
	if t, ok := ctx.Deadline(); ok {
		t = t.Add(-time.Second)
		if t.Before(stopThinking) {
			stopThinking = t
		}
	}

	// Evaluate the board and find the strongest guarantee we can ensure.
	moves := evaluateBoard(b, mark)
	strongest := strongest(moves).Guarantee

	// Pick a random strong move.
	var strongMoves []ttt.Move
	for _, m := range moves {
		if m.Guarantee == strongest {
			strongMoves = append(strongMoves, m.Move)
		}
	}
	move := strongMoves[p.rand.Intn(len(strongMoves))]

	// Take some more time to pretend we're still "thinking".
	time.Sleep(time.Until(stopThinking))

	return move, nil
}

type guarantee uint8

const (
	guaranteeLoss guarantee = iota
	guaranteeTie
	guaranteeWin
)

type evaluatedMove struct {
	Move      ttt.Move
	Guarantee guarantee
}

func evaluateBoard(b ttt.Board, mark ttt.State) []evaluatedMove {
	legalMoves := legalMoves(b)

	// Fast path for empty board.
	if len(legalMoves) == len(b.Cells) {
		// All first moves are known to only guarantee a tie.
		var moves []evaluatedMove
		for _, move := range legalMoves {
			moves = append(moves, evaluatedMove{
				Move:      move,
				Guarantee: guaranteeTie,
			})
		}
		return moves
	}

	var moves []evaluatedMove
	for _, move := range legalMoves {
		moves = append(moves, evaluatedMove{
			Move:      move,
			Guarantee: evaluateMove(move, b, mark),
		})
	}
	return moves
}

func evaluateMove(move ttt.Move, b ttt.Board, mark ttt.State) guarantee {
	err := b.Apply(move, mark)
	if err != nil {
		panic(fmt.Errorf("internal error: error applying a move: %v", err))
	}
	switch b.Condition() {
	case ttt.XWon:
		if mark == ttt.X {
			return guaranteeWin
		} else {
			return guaranteeLoss
		}
	case ttt.OWon:
		if mark == ttt.O {
			return guaranteeWin
		} else {
			return guaranteeLoss
		}
	case ttt.Tie:
		return guaranteeTie
	case ttt.NotEnd:
		// See what would happen if the opponent plays perfectly
		// and makes a follow-up move with the strongest guarantee.
		opponent := opponentOf(mark)
		opponentMoves := evaluateBoard(b, opponent)
		strongestOpponentMove := strongest(opponentMoves).Move
		switch evaluateMove(strongestOpponentMove, b, opponent) {
		case guaranteeLoss:
			// If the strongest opponent follow-up move can only guarantee them a loss,
			// that means our move can guarantee a win.
			return guaranteeWin
		case guaranteeTie:
			// If the strongest opponent follow-up move can only guarantee them a tie,
			// that means our move can only guarantee a tie.
			return guaranteeTie
		case guaranteeWin:
			// If the strongest opponent follow-up move can guarantee them a win,
			// that means our move can only guarantee a loss.
			return guaranteeLoss
		default:
			panic("unreachable")
		}
	default:
		panic("unreachable")
	}
}

// legalMoves returns all legal moves on board b.
func legalMoves(b ttt.Board) []ttt.Move {
	var moves []ttt.Move
	for i, cell := range b.Cells {
		if cell != ttt.F {
			continue
		}
		moves = append(moves, ttt.Move(i))
	}
	return moves
}

// strongest returns a move with the strongest guarantee.
// moves must contain at least 1 element.
func strongest(moves []evaluatedMove) evaluatedMove {
	strongest := moves[0]
	for _, m := range moves[1:] {
		if m.Guarantee > strongest.Guarantee {
			strongest = m
		}
	}
	return strongest
}

func opponentOf(mark ttt.State) ttt.State {
	switch mark {
	case ttt.X:
		return ttt.O
	case ttt.O:
		return ttt.X
	default:
		panic("unreachable")
	}
}
