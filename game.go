package main

import "github.com/corentings/chess/v2"

type ViewMode int

const (
	ViewLive ViewMode = iota
	ViewHistory
)

type Game struct {
	LiveGame *chess.Game

	ViewPos  *chess.Position
	ViewMode ViewMode

	Moves *MoveList
	Clock ChessClock
	AI    *AIPlayer

	Input *InputState

	Selected *chess.Square
	Legal    map[chess.Square]bool

	Promotion *PromotionState

	Mode   GameMode
	Over   bool
	Result string
}

func (g *Game) Update() error {
	switch g.Mode {
	case ModeMenu:
		g.MenuUpdate()
		return nil
	case ModePlaying:
		return g.updatePlaying()
	}
	return nil
}

func (g *Game) updatePlaying() error {
	g.Input.Update()

	// 1. History navigation
	g.Moves.UpdateHistory()

	if g.Moves.ViewingHistory() {
		g.ViewMode = ViewHistory
		g.ViewPos = BuildPositionAt(g.LiveGame, g.Moves.ViewIndex)
	} else {
		g.ViewMode = ViewLive
		g.ViewPos = nil
	}

	// 2. Clock (only when live)
	if g.ViewMode == ViewLive && g.Clock != nil {
		g.Clock.Update(g.LiveGame.Position().Turn())
	}

	// 3. Promotion overrides everything
	if g.Promotion.Active {
		g.Promotion.Update(g)
		return nil
	}

	// 4. AI move
	if g.ViewMode == ViewLive && g.AI.ShouldMove(g.LiveGame) {
		g.AI.TryMove(g)
		return nil
	}

	// 5. Human input
	if g.ViewMode == ViewLive {
		g.HandleHumanInput()
	}

	// 6. End game
	g.CheckGameOver()

	return nil
}
