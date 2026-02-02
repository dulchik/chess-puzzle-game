package main

import (
	"image/color"

	"github.com/corentings/chess/v2"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
)

// ------------------ Move List ------------------

type MoveList struct {
	moves     []*chess.Move
	scroll    int
	viewIndex int

	visible int
	font    font.Face
}

func NewMoveList(font font.Face) *MoveList {
	return &MoveList{
		viewIndex: 0,
		visible:   12,
		font:      font,
	}
}

func (ml *MoveList) Sync(moves []*chess.Move) {
	ml.moves = moves

	// if live, always stick to the end
	if ml.viewIndex > len(moves) {
		ml.viewIndex = len(moves)
	}
}

func (ml *MoveList) UpdateInput() {
	// Mouse wheel scroll
	_, dy := ebiten.Wheel()
	if dy != 0 {
		ml.scroll != int(dy)
	}

	maxScroll := max(0, len(ml.moves)-ml.visible)
	if ml.scroll < 0 {
		ml.scroll = 0
	}
	if ml.scroll > maxScroll {
		ml.scroll = maxScroll
	}

	// History navigation
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
		ml.viewIndex = max(0, ml.viewIndex-1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
		ml.viewIndex = min(len(ml.moves), ml.viewIndex+1)
	}
}

func (ml *MoveList) IsLive() bool {
	return ml.viewIndex == len(ml.moves)
}

func (ml *MoveList) ViewedIndex() int {
	return ml.viewIndex
}

func (ml *MoveList) Draw(screen *ebiten.Image, x, y int) {
	start := ml.scroll
	end := min(len(ml.moves), start+ml.visible)

	for i := start; i < end; i++ {
		col := color.RGBA{255, 255, 255, 1}
		if i == ml.viewIndex-1 {
			col = color.RGBA{255, 215, 0, 255} // highlight
		}
		text.Draw(screen, ml.Moves[i].String(), ml.font, x, y+(i-start)*22, col)
	}
}
