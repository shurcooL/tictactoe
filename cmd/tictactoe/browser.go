// +build js

package main

import (
	"fmt"
	"html/template"
	"syscall/js"

	"github.com/shurcooL/htmlg"
	ttt "github.com/shurcooL/tictactoe"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"honnef.co/go/js/dom/v2"
)

var document = dom.GetWindow().Document().(dom.HTMLDocument)

func displayGameStart(board ttt.Board, players [2]player, cellClick chan<- int) {
	// Wait for DOM to finish loading.
	waitDOM()

	document.SetTitle("Tic-Tac-Toe")

	// When a board cell is clicked, send its [0, 9) index to cellClick channel.
	js.Global().Set("CellClick", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		index := args[0].Int()
		select {
		case cellClick <- index:
		default:
		}
		return nil
	}))
}

func displayTurnStart(board ttt.Board, players [2]player, active player, condition ttt.Condition) {
	// Draw page at start of turn.
	_, isCellClicker := active.Player.(ttt.CellClicker)
	document.Body().SetInnerHTML(htmlg.Render(page{Board: board, Turn: active.Mark, Clickable: isCellClicker, Condition: condition, Players: players}.Render()...))
}

func displayTurnEnding(board ttt.Board, players [2]player, condition ttt.Condition) {
	// Draw page after player finished turn.
	document.Body().SetInnerHTML(htmlg.Render(page{Board: board, Condition: condition, Players: players}.Render()...))
}

func displayGameEnd(board ttt.Board, players [2]player, condition ttt.Condition) {
	// Draw page at end of game.
	document.Body().SetInnerHTML(htmlg.Render(page{Board: board, Condition: condition, Players: players}.Render()...))
	select {}
}

func displayError(board ttt.Board, players [2]player, err error) {
	// Draw page on error.
	document.Body().SetInnerHTML(htmlg.Render(page{Board: board, ErrorMessage: err.Error(), Players: players}.Render()...))
	select {}
}

// page renders the entire page body.
type page struct {
	Board        ttt.Board
	Turn         ttt.State
	Clickable    bool
	Condition    ttt.Condition
	ErrorMessage string
	Players      [2]player
}

func (p page) Render() []*html.Node {
	var statusMessage *html.Node
	switch {
	case p.ErrorMessage != "":
		statusMessage = style(
			`line-height: 60px; text-align: center; color: red;`,
			htmlg.Div(htmlg.Text(p.ErrorMessage)),
		)
	case p.Condition != ttt.NotEnd:
		statusMessage = style(
			`line-height: 60px; text-align: center;`,
			htmlg.Div(htmlg.Text(p.Condition.String())),
		)
	default:
		statusMessage = style(`height: 60px;`, htmlg.Div())
	}
	return []*html.Node{
		style(
			`text-align: center; margin-top: 50px;`,
			htmlg.Div(
				// Player X.
				style(
					`display: inline-block; width: 200px;`,
					htmlg.Span(p.Players[0].Render(p.Turn)...),
				),
				// Board.
				style(
					`display: inline-block; margin-left: 30px; margin-right: 30px;`,
					htmlg.Span(board{Board: p.Board, Clickable: p.Clickable}.Render()...),
				),
				// Player O.
				style(
					`display: inline-block; width: 200px;`,
					htmlg.Span(p.Players[1].Render(p.Turn)...),
				),
			),
		),
		statusMessage,
		// Give credit to Renee French for the Go gopher.
		style(
			`text-align: right; font-style: italic;`,
			htmlg.Div(htmlg.Text("Go gopher by Renee French.")),
		),
	}
}

// board renders a board.
type board struct {
	ttt.Board
	Clickable bool
}

func (b board) Render() []*html.Node {
	table := &html.Node{Data: atom.Table.String(), Type: html.ElementNode}
	for row := 0; row < 3; row++ {
		tr := &html.Node{Data: atom.Tr.String(), Type: html.ElementNode}
		for col, cell := range b.Cells[3*row : 3*row+3] {
			td := &html.Node{Data: atom.Td.String(), Type: html.ElementNode}
			htmlg.AppendChildren(td, boardCell{State: cell, Clickable: b.Clickable, Index: 3*row + col}.Render()...)
			tr.AppendChild(td)
		}
		table.AppendChild(tr)
	}
	return []*html.Node{
		table,
	}
}

// boardCell renders a board cell.
type boardCell struct {
	ttt.State
	Clickable bool
	Index     int
}

func (c boardCell) Render() []*html.Node {
	cell := style(
		`display: table-cell; width: 30px; height: 30px; text-align: center; vertical-align: middle; background-color: #f4f4f4;`,
		htmlg.Div(
			htmlg.Text(c.String()),
		),
	)
	if c.Clickable {
		cell = &html.Node{
			Type: html.ElementNode, Data: atom.A.String(),
			Attr: []html.Attribute{
				{Key: atom.Style.String(), Val: `display: block; cursor: pointer;`},
				{Key: atom.Onclick.String(), Val: fmt.Sprintf(`CellClick(%d);`, c.Index)},
			},
			FirstChild: cell,
		}
	}
	return []*html.Node{cell}
}

// Render the player. turn indicates whose turn it currently is.
func (p player) Render(turn ttt.State) []*html.Node {
	switch imager, ok := p.Player.(ttt.Imager); ok {
	case true:
		var imgStyle string
		switch p.Mark {
		case ttt.X:
			imgStyle = `height: 100px;`
		case ttt.O:
			imgStyle = `height: 100px; transform: scaleX(-1);`
		}
		text := htmlg.Text(fmt.Sprintf("%v (%v)", p.Name(), p.Mark))
		if p.Mark == turn {
			text = &html.Node{
				Type: html.ElementNode, Data: atom.Strong.String(),
				FirstChild: text,
			}
		}
		return []*html.Node{
			style(
				imgStyle,
				img(imager.Image()),
			),
			htmlg.Div(text),
		}
	case false:
		text := htmlg.Text(fmt.Sprintf("%v (%v)", p.Name(), p.Mark))
		if p.Mark == turn {
			text = &html.Node{
				Type: html.ElementNode, Data: atom.Strong.String(),
				FirstChild: text,
			}
		}
		return []*html.Node{
			text,
		}
	default:
		panic("unreachable")
	}
}

// img returns an image element <img src="{{.src}}">.
func img(src template.URL) *html.Node {
	img := &html.Node{
		Type: html.ElementNode, Data: atom.Img.String(),
		Attr: []html.Attribute{{Key: atom.Src.String(), Val: string(src)}},
	}
	return img
}

func style(style string, n *html.Node) *html.Node {
	if n.Type != html.ElementNode {
		panic("invalid node type")
	}
	n.Attr = append(n.Attr, html.Attribute{Key: atom.Style.String(), Val: style})
	return n
}

func waitDOM() {
	if document.ReadyState() != "loading" {
		// Already loaded.
		return
	}
	var ch = make(chan struct{})
	document.AddEventListener("DOMContentLoaded", false, func(dom.Event) { close(ch) })
	<-ch
}
