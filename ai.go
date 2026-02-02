package main

import (
	"github.com/corentings/chess/v2"
)

type AIController struct {
	engine   *Engine
	color    chess.Color
	thinking bool
}

func NewAIController(engine *Engine, color chess.Color) *AIController {
	return &AIController{
		engine: engine,
		color:  color,
	}
}

func (ai *AIController) IsAITurn(pos *chess.Position) bool {
	return ai.engine != nil && pos.Turn() == ai.color
}

func (ai *AIController) IsThinking() bool {
	return ai.thinking
}

func (ai *AIController) Update(g *Game) {
	if ai.engine == nil || ai.thinking || g.gameOver {
		return
	}

	pos := g.liveGame.Position()

	// Do not move during history view
	if !g.moveList.IsLive() {
		return
	}

	if pos.Turn() != ai.color {
		return
	}

	ai.thinking = true

	go func() {
		fen := pos.String()
		uci, err := ai.engine.BestMove(fen, 300)
		if err == nil {
			move, err := chess.UCINotation{}.Decode(g.liveGame.Position(), uci)
			if err == nil {
				g.liveGame.Move(move, nil)
			}
		}

		ai.thinking = false
	}()
}
