package main

import (
	"image/color"
	"io/ioutil"
	"log"
	"github.com/golang/freetype/truetype"

	"github.com/corentings/chess/v2"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
)

const (
	screenWidth  = 1280
	screenHeight = 1280
	tileSize     = 160
)

var boardColors = [2]color.Color{
	color.RGBA{240, 217, 181, 255}, // light
	color.RGBA{181, 136, 99, 255},  // dark
}

type Game struct {
	chessGame *chess.Game
}

func (g *Game) Update() error {
	// TODO: handle mouse clicks and move pieces
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	fontBytes, _ := ioutil.ReadFile("chess_merida_unicode.ttf")
	f, _ := truetype.Parse(fontBytes)

	opts := truetype.Options{}
	opts.Size = 128 

	optsFont := truetype.Options{}
	optsFont.Size = 58 

	facePieces := truetype.NewFace(f, &opts)
	faceFont := truetype.NewFace(f, &optsFont)

	// Draw chessboard
	for rank := chess.Rank8; rank >= chess.Rank1; rank-- {
		for file := chess.FileA; file <= chess.FileH; file++ {
			sq := chess.Square(int(file) + 8*int(rank))
			col := boardColors[(int(rank)+int(file))%2]
			ebitenutil.DrawRect(screen, float64(file)*tileSize, float64(7-rank)*tileSize, tileSize, tileSize, col)

			// Draw piece as string for now
			piece := g.chessGame.Position().Board().Piece(sq)
			if piece != chess.NoPiece {
				ebitenutil.DebugPrintAt(screen, piece.String(), int(file)*tileSize+20, int(7-rank)*tileSize+20)
			}
			text.Draw(screen, piece.String(), facePieces, int(file)*tileSize+16, int(7-rank)*tileSize+128, color.Black)
		}
	}

	// Draw rank and file notations
	for rank := chess.Rank8; rank >= chess.Rank1; rank-- {
		colInv := boardColors[1-((int(rank)+int(chess.FileA))%2)]
		y := int(7-rank)*tileSize+40 
 	   	text.Draw(
    	    screen,
    	    rank.String(),
    	    faceFont,
    	    -15,
			y,
			colInv,
   	 	)
	}
	for file := chess.FileA; file <= chess.FileH; file++ {
		colInv := boardColors[1-((int(chess.Rank1)+int(file))%2)]
   	 	text.Draw(
   	    	screen,
    		file.String(),
    	   	faceFont,
     	 	int(file)*tileSize+108,
			1280, // bottom margin
      	 	colInv,
		)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	game := &Game{chessGame: chess.NewGame()}
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Ebiten Chess")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
