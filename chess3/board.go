package main


import (
	"image/color"

	"github.com/corentings/chess/v2"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
)

type BoardRenderer struct {
	tileSize 	int
	light 		color.Color
	dark		color.Color

	pieceFont	font.Face
	uiFont		font.Face
}

func (g *Game) Draw(screen *ebiten.Image) {
	pos := g.LiveGame.Position()
	if g.ViewMode == ViewHistory {
		pos = g.ViewPos
	}

	DrawBoard(screen, pos, g)
	DrawMoveList(screen, g.Moves)
	DrawClock(screen, g.Clock)

	if g.Promotion.Active {
		g.Promotion.Draw(screen)
	}
}
