package main

import (
    "image/color"
    "log"

    "github.com/corentings/chess/v2"
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
    screenWidth  = 480
    screenHeight = 480
    tileSize     = 60
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
        }
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

