package main

import (
	"github.com/corentings/chess/v2"
	"github.com/hajimehoshi/ebiten/v2"
)

type InputHandler struct {
	tileSize int
	mouseDown bool
}

func NewInputHandler(tileSize int) *InputHandler {
	return &InputHandler{tileSize: tileSize}
}

func (ih *InputHandler) Update(g *Game) {
	mousePressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)

	// Promotion Picker
	if g.promotionFrom != nil {
		ih.handlePromotion(g, mousePressed)
		ih.mouseDown = mousePressed
		return
	}

	// Ignore input when game is over or history view
	if g.gameOver || !g.moveList.IsLive() {
		ih.mouseDown = mousePressed
		return
	}

	// Ignore if AI turn
	if g.liveGame.Position().Turn() == g.aiColor {
		ih.mouseDown = mousePressed
		return
	}
	
	// Detect click (not hold) 
	if mousePressed && !ih.mouseDown {
		x, y := ebiten.CursorPosition()
		sq, ok := ih.squareFromMouse(x, y)
		if ok {
			ih.handleBoardClick(g, sq)
		}

	}

	ih.mouseDown = mousePressed
}

func (ih *InputHandler) handleBoardClick(g *Game, sq chess.Square) {
	board := g.liveGame.Position().Board()
	piece := board.Piece(sq)

	// No selection yet
	if g.selectedSquare == nil {
		if piece != chess.NoPiece && piece.Color() == g.liveGame.Position().Turn() {
			g.selectedSquare = &sq
			g.legalTargets = legalMovesFrom(g.liveGame.Position(), sq)
		}
		return
	}

	// Try pawn promotion
	if isPawnPromotion(g.liveGame.Position(), *g.selectedSquare, sq) {
		g.promotionFrom = g.selectedSquare
		g.promotionTo = &sq
		g.selectedSquare = nil
		g.legalTargets = nil
		return
	}

	// Try normal move
	move, err := moveFromSquare(g.liveGame.Position(), *g.selectedSquare, sq)
	if err == nil {
		g.liveGame.Move(move, nil)
		g.moveList.Sync(g.liveGame.Moves())
	}

	g.selectedSquare = nil
	g.legalTargets = nil
}

func (ih *InputHandler) handlePromotion(g *Game, mousePressed bool) {
	// Right click cancels
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) && !ih.mouseDown {
		g.promotionFrom = nil
		g.promotionTo = nil
		return
	}

	if !mousePressed || ih.mouseDown {
		return
	}

	x, y := ebiten.CursorPosition()
	file := x / ih.tileSize
	rank := 7 - (y / ih.tileSize)

	options := []rune{'q', 'r', 'b', 'n'}

	for i, promo := range options {
		r := g.promotionTo.Rank()
		if g.liveGame.Position().Turn() == chess.Black {
			r += chess.Rank(i)
		} else {
			r -= chess.Rank(i)
		}

		if int(file) == int(g.promotionTo.File()) && int(rank) == int(r) {
			uci := g.promotionFrom.String() + g.promotionTo.String() + string(promo)
			move, err := chess.UCINotation{}.Decode(g.liveGame.Position(), uci)
			if err == nil {
				g.liveGame.Move(move, nil)
				g.moveList.Sync(g.liveGame.Moves())
			}

			g.promotionFrom = nil
			g.promotionTo = nil
			return
		}
	}

	g.promotionFrom = nil
	g.promotionTo = nil
}

func (ih *InputHandler) squareFromMouse(x, y int) (chess.Square, bool) {
	if x < 0 || y < 0 || x >= ih.tileSize*8 || y >= ih.tileSize*8 {
		return 0, false
	}

	file := x / ih.tileSize
	rank := 7 - (y / ih.tileSize)

	if file < 0 || file > 7 || rank < 0 || rank > 7 {
		return 0, false
	}

	return chess.Square(file + 8*rank), true
}


