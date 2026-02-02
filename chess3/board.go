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

func NewBoardRenderer(tileSize int, pieceFont, uiFont font.Face) *BoardRenderer {
	return &BoardRenderer{
		tileSize: 	tileSize,
		light:		color.RGBA{240, 217, 181, 255},
		dark:		color.RGBA{181, 136, 99, 255},
		pieceFont: 	pieceFont,
		uiFont:		uiFont,
	}
}

func (br *BoardRenderer) Draw(
	screen *ebiten.Image,
	pos *chess.Position,
	selected *chess.Square,
	legalTargets map[chess.Square]bool,
	lastMove *chess.Move,
	kingInCheck bool,
) {
	for rank := chess.Rank8; rank >= chess.Rank1; rank-- {
		for file := chess.FileA; file <= chess.FileH; file++ {
			sq := chess.Square(int(file) + 8*int(rank))

			x := float64(int(file) * br.tileSize)
			y := float64((7-int(rank)) * br.tileSize)

			// square color
			col := br.light
			if (int(rank)+int(file))%2 == 1 {
				col = br.dark
			}
			ebitenutil.DrawRect(screen, x, y, float64(br.tileSize), float64(br.tileSize), col)

			// selected square
			if selected != nil && *selected == sq {
			ebitenutil.DrawRect(screen, x, y, float64(br.tileSize), float64(br.tileSize), color.RGBA{0, 255, 0, 80})
			}


			// legal targets
			if legalTargets != nil && legalTargets[sq] {
			ebitenutil.DrawRect(screen, x, y, float64(br.tileSize), float64(br.tileSize), color.RGBA{0, 0, 0, 60})
			}


			// last move highlight
			if lastMove != nil && (sq == lastMove.S1() || sq == lastMove.S2()) {
			ebitenutil.DrawRect(screen, x, y, float64(br.tileSize), float64(br.tileSize), color.RGBA{255, 255, 0, 80})
			}


			// king in check highlight
			if kingInCheck {
				piece := pos.Board().Piece(sq)
				if piece != chess.NoPiece && piece.Type() == chess.King && piece.Color() == pos.Turn() {
					ebitenutil.DrawRect(screen, x, y, float64(br.tileSize), float64(br.tileSize), color.RGBA{200, 0, 0, 120})
				}
			}

			// draw piece 
			piece := pos.Board().Piece(sq)
			if piece != chess.NoPiece {
				text.Draw(screen, piece.String(), br.pieceFont, int(x)+16, int(y)+br.tileSize-32, color.Black)
			}
		}
	}

	br.drawCoordinates(screen)
}

func (br *BoardRenderer) drawCoordinates(screen *ebiten.Image) {
	for rank := chess.Rank8; rank >= chess.Rank1; rank-- {
		y := (7-int(rank))*br.tileSize + 24
		text.Draw(screen, rank.String(), br.uiFont, 4, y, color.Black)
	}
	
	for file := chess.FileA; file <= chess.FileH; file++ {
		x := int(file)*br.tileSize + br.tileSize - 20
		y := br.tileSize*8 - 4

		text.Draw(screen, file.String(), br.uiFont, x, y, color.Black)
	}
}
