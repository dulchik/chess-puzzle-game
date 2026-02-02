package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	screenWidth  = 1680
	screenHeight = 1280
)

type App struct {
	menu  *menu.Menu
	game  *game.Game
	board *board.Board
	input *input.Controller
	ai    *ai.Controller
}

func NewApp() *App {
	// --- Engine ---
	eng, err := engine.New("./engine/stockfish/stockfish-ubuntu-x86-64-bmi2")
	if err != nil {
		log.Fatal(err)
	}

	// --- Core systems ---
	menuSystem := menu.New()
	gameSystem := game.New()
	boardSystem := board.New()
	inputSystem := input.New()
	aiSystem := ai.New(eng)

	return &App{
		menu:  menuSystem,
		game:  gameSystem,
		board: boardSystem,
		input: inputSystem,
		ai:    aiSystem,
	}
}

func (a *App) Update() error {
	// ---- MENU MODE ----
	if a.menu.Active {
		a.menu.Update()

		if a.menu.StartRequested {
			cfg := a.menu.ConsumeConfig()

			a.game.Start(cfg)
			a.ai.Configure(cfg)
			a.menu.Active = false
		}

		return nil
	}

	// ---- GAME MODE ----
	a.input.Update()
	a.game.Update(a.input)

	if a.ai.Enabled() {
		a.ai.Update(a.game)
	}

	return nil
}

func (a *App) Draw(screen *ebiten.Image) {
	if a.menu.Active {
		a.menu.Draw(screen)
		return
	}

	a.board.Draw(screen, a.game.View())
	a.game.DrawUI(screen)
}

func (a *App) Layout(_, _ int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	app := NewApp()

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Chess")

	if err := ebiten.RunGame(app); err != nil {
		log.Fatal(err)
	}
}
