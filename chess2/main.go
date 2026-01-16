package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
)

func main() {
	const chessboardSquaresPerSide = 8
	const chessboardSquares = chessboardSquaresPerSide * chessboardSquaresPerSide

	const chessboardSquareSideSize = 120
	const chessboardSideSize = chessboardSquaresPerSide * chessboardSquareSideSize
	
	chessboardRect := image.Rect(0, 0, chessboardSideSize, chessboardSideSize)
	chessboardImage := image.NewRGBA(chessboardRect)
	uniform := &image.Uniform{}


	for s := 0; s < chessboardSquares; s++ {
		x := s / chessboardSquaresPerSide
		y := s % chessboardSquaresPerSide

		x0 := x * chessboardSquareSideSize
		y0 := y * chessboardSquareSideSize
		x1 := x0 + chessboardSquareSideSize
		y1 := y0 + chessboardSquareSideSize
		
		rect := image.Rect(x0, y0, x1, y1)

		if isBlack := (x+y)%2 == 1; isBlack {
			uniform.C = color.RGBA{R: 184, G: 135, B: 98, A: 255}
		} else {
			uniform.C = color.RGBA{R: 237, G: 214, B: 176, A: 255}
		}

		draw.Draw(chessboardImage, rect, uniform, image.Point{}, draw.Src)

	}

	chessboardFile, err := os.Create("chessboard.png")
	if err != nil {
		panic(err)
	}


	if err := png.Encode(chessboardFile, chessboardImage); err != nil {
		chessboardFile.Close()
		panic(err)
	}

	if err := chessboardFile.Close(); err != nil {
		panic(err)
	}
}
