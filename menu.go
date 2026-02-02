package main

import (
	"image/color"

	"github.com/corentings/chess/v2"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type MenuMode int

const (
	MenuMain MenuMode = iota
	MenuAIOptions
)

type Menu struct {
	Active bool

	Mode MenuMode

	PlayWithAI   bool
	PlayerColor  chess.Color
	AIElo        int
	TimerSeconds int

	cursor int
}

func NewMenu() *Menu {
	return &Menu{
		Mode:         MenuMain,
		PlayWithAI:   false,
		PlayerColor:  chess.White,
		AIElo:        800,
		TimerSeconds: 300,
	}
}

func (m *Menu) Update() {
	if !m.Active {
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		m.cursor++
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		m.cursor--
	}

	if m.cursor < 0 {
		m.cursor = 0
	}

	switch m.Mode {
	case MenuMain:
		if m.cursor > 2 {
			m.cursor = 2
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			switch m.cursor {
			case 0:
				m.PlayWithAI = false
				m.Active = false
			case 1:
				m.PlayWithAI = true
				m.Mode = MenuAIOptions
				m.cursor = 0
			case 2:
				ebiten.Exit()
			}
		}

	case MenuAIOptions:
		if m.cursor > 3 {
			m.cursor = 3
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
			m.adjust(-1)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
			m.adjust(1)
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			m.Active = false
		}
	}
}

func (m *Menu) adjust(dir int) {
	switch m.cursor {
	case 0:
		if m.PlayerColor == chess.White {
			m.PlayerColor = chess.Black
		} else {
			m.PlayerColor = chess.White
		}
	case 1:
		m.AIElo += dir * 100
		if m.AIElo < 400 {
			m.AIElo = 400
		}
		if m.AIElo > 2400 {
			m.AIElo = 2400
		}
	case 2:
		m.TimerSeconds += dir * 60
		if m.TimerSeconds < 0 {
			m.TimerSeconds = 0
		}
	}
}

func (m *Menu) Draw(screen *ebiten.Image) {
	if !m.Active {
		return
	}

	bg := color.RGBA{20, 20, 20, 220}
	screen.Fill(bg)

	y := 200
	drawItem := func(text string, selected bool) {
		col := color.White
		if selected {
			col = color.RGBA{255, 200, 0, 255}
		}
		DrawText(screen, text, 200, y, col)
		y += 40
	}

	switch m.Mode {
	case MenuMain:
		drawItem("Local Multiplayer", m.cursor == 0)
		drawItem("Play vs AI", m.cursor == 1)
		drawItem("Quit", m.cursor == 2)

	case MenuAIOptions:
		drawItem("Color: "+m.PlayerColor.String(), m.cursor == 0)
		drawItem("AI Elo: "+itoa(m.AIElo), m.cursor == 1)
		drawItem("Timer: "+itoa(m.TimerSeconds/60)+" min", m.cursor == 2)
		drawItem("Start Game", m.cursor == 3)
	}
}
