package main

import (
	"image/color"
	"log"
	"os"

	"github.com/corentings/chess/v2"
	"github.com/corentings/chess/v2/image"
)

func main() {
    // create file
    f, err := os.Create("example.svg")
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()

    // create board position
    fenStr := "rnbqkbnr/pppppppp/8/8/3P4/8/PPP1PPPP/RNBQKBNR b KQkq - 0 1"
    pos := &chess.Position{}
    if err := pos.UnmarshalText([]byte(fenStr)); err != nil {
        log.Fatal(err)
    }

    // write board SVG to file
    yellow := color.RGBA{255, 255, 0, 1}
    mark := image.MarkSquares(yellow, chess.D2, chess.D4)
    if err := image.SVG(f, pos.Board(), mark); err != nil {
        log.Fatal(err)
    }
}
