package main

import (
	"fmt"
	"image/color"
	"io/ioutil"
	"log"
	"time"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"github.com/corentings/chess/v2"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
)

// ---------------- TIMER ------------------

type ChessClock struct {
	White		time.Duration
	Black		time.Duration
	last		time.Time
	lastTurn 	chess.Color
	Enabled 	bool
}

func NewClock(seconds int) ChessClock {
	t := time.Duration(seconds) * time.Second
	return ChessClock{
		White:	 t,
		Black:	 t,
		last:    time.Now(),
		Enabled: true,
	}
}

func (c *ChessClock) Update(turn chess.Color) {
	if !c.Enabled {
		return
	}

	now := time.Now()

	if c.last.IsZero() || c.lastTurn != turn {
		c.last = now
		c.lastTurn = turn
		return
	}

	delta := now.Sub(c.last)
	c.last = now

	if turn == chess.White {
		c.White -= delta
	} else {
		c.Black -= delta
	}
}

func (c *ChessClock) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("White: %s", formatTime(c.White)), 20, 20)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Black: %s", formatTime(c.Black)), 20, 40)
}

func formatTime(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}

type GameMode int

const (
	ModeMenu GameMode = iota
	ModePlaying
)

const (
	tileSize     = 160
	boardSize    = tileSize * 8 // 1280
	panelWidth   = 400
	screenWidth  = boardSize + panelWidth
	screenHeight = boardSize

)

var (
	pieceFace font.Face
	textFace  font.Face
)

var boardColors = [2]color.Color{
	color.RGBA{240, 217, 181, 255}, // light
	color.RGBA{181, 136, 99, 255},  // dark
}

type Game struct {
	chessGame 		*chess.Game
	selectedSquare 	*chess.Square
	mouseDown		bool
	legalTargets	map[chess.Square]bool

	// promotion UI
	promotionFrom   *chess.Square
	promotionTo     *chess.Square

	gameOver 		bool
	gameResult   	string

	engine 			*Engine
	aiColor			chess.Color
	aiElo			int
	aiThinking		bool

	mode 			GameMode

	// menu settings
	playWithAI		bool
	playerColor 	chess.Color
	timeSeconds		int

	useTimer 		bool
	clock			ChessClock

	whiteTime time.Duration
	blackTime time.Duration
	lastTick  time.Time
}

func loadFonts() {
	// Chess pieces
	pieceBytes, _ := ioutil.ReadFile("chess_merida_unicode.ttf")
	pieceTTF, _ := truetype.Parse(pieceBytes)
	pieceFace = truetype.NewFace(pieceTTF, &truetype.Options{Size: 128})

	// UI text
	textBytes, _ := ioutil.ReadFile("Roboto-Regular.ttf")
	textTTF, _ := truetype.Parse(textBytes)
	textFace = truetype.NewFace(textTTF, &truetype.Options{Size: 20})
}

// Moves Draw helpers

func legalMovesFrom(pos *chess.Position, from chess.Square) map[chess.Square]bool {
	moves := pos.ValidMoves()
	targets := make(map[chess.Square]bool)

	for _, m := range moves {
		if m.S1() == from {
			targets[m.S2()] = true
		}
	}
	return targets
}

func squareFromMouse(x, y int) (chess.Square, bool) {
	if x < 0 || y < 0 || x >= screenWidth || y >= screenHeight {
		return 0, false
	}

	file := x / tileSize
	rank := 7 - (y / tileSize)

	if file < 0 || file > 7 || rank < 0 || rank > 7 {
		return 0, false
	} 

	return chess.Square(file + 8*rank), true
}

func moveFromSquares(pos *chess.Position, from, to chess.Square) (*chess.Move, error) {
	uci := from.String() + to.String()
	notation := chess.UCINotation{}
	return notation.Decode(pos, uci)
}

func formatMoves(moves []*chess.Move) []string {
	var lines []string

	for i := 0; i < len(moves); i += 2 {
		moveNum := i/2 + 1
		line := fmt.Sprintf("%d. %s", moveNum, moves[i].String())

		if i+1 < len(moves) {
			line += " " + moves[i+1].String()
		}

		lines = append(lines, line)
	}
	return lines
}

// Game Update helpers

func (g *Game) resetGame() {
	g.chessGame = chess.NewGame()

	g.selectedSquare = nil
	g.legalTargets = nil

	g.promotionFrom = nil
	g.promotionTo = nil

	g.gameOver = false
}

func isPawnPromotion(pos *chess.Position, from, to chess.Square) bool {
    piece := pos.Board().Piece(from)
    if piece == chess.NoPiece || piece.Type() != chess.Pawn {
        return false
    }

    if piece.Color() == chess.White && to.Rank() == chess.Rank8 {
        return true
    }
    if piece.Color() == chess.Black && to.Rank() == chess.Rank1 {
        return true
    }
    return false
}

func (g *Game) handleHumanClick(sq chess.Square) {
	board := g.chessGame.Position().Board()
	piece := board.Piece(sq)

	if g.selectedSquare == nil {
		if piece != chess.NoPiece && piece.Color() == g.chessGame.Position().Turn() {
			g.selectedSquare = &sq
			g.legalTargets = legalMovesFrom(g.chessGame.Position(), sq)
		}
		return
	}

	// Try move
	if isPawnPromotion(g.chessGame.Position(), *g.selectedSquare, sq) {
		g.promotionFrom = g.selectedSquare
		g.promotionTo = &sq
		g.selectedSquare = nil
		g.legalTargets = nil
		return
	}
	uci := g.selectedSquare.String() + sq.String()
	move, err := chess.UCINotation{}.Decode(g.chessGame.Position(), uci)
	if err == nil {
 		g.chessGame.Move(move, nil) 
	} 
	
	g.selectedSquare = nil
	g.legalTargets = nil
}

func (g *Game) handlePromotion(mousePressed bool) {
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) && !g.mouseDown {
		g.promotionFrom = nil
		g.promotionTo = nil
		return
	}

	if !mousePressed || g.mouseDown {
		return
	}

	x, y := ebiten.CursorPosition()
	file := x / tileSize
	rank := 7 - (y / tileSize)

	options := []rune{'q', 'r', 'b', 'n'}

	for i, promo := range options {
		r := g.promotionTo.Rank()

		if g.chessGame.Position().Turn() == chess.Black {
			r += chess.Rank(i)	
		} else {
			r -= chess.Rank(i)
		}

		if int(file) == int(g.promotionTo.File()) && int(rank) == int(r) {
			uci := g.promotionFrom.String() + g.promotionTo.String() + string(promo)
			move, err := chess.UCINotation{}.Decode(g.chessGame.Position(), uci)
			if err == nil {
				g.chessGame.Move(move, nil)
			}

			g.promotionFrom = nil
			g.promotionTo = nil
			return
		}
	}
	
	g.promotionFrom = nil
	g.promotionTo = nil
}

func (g *Game) tryAIMove() {
	if g.aiThinking {
		return
	}

	g.aiThinking = true

	go func() {
		fen := g.chessGame.Position().String()
		uci, err := g.engine.BestMove(fen, 300)
		if err == nil {
			move, err := chess.UCINotation{}.Decode(g.chessGame.Position(), uci)
			if err == nil {
				g.chessGame.Move(move, nil)
			}
		}
		g.aiThinking = false
	}()

}

func (g *Game) checkGameOver() {
	pos := g.chessGame.Position()
	switch pos.Status() {
	case chess.Checkmate:
		g.gameOver = true
		if pos.Turn() == chess.White {
			g.gameResult = "Black wins by checkmate"
		} else {
			g.gameResult = "White wins by checkmate"
		}
	case chess.Stalemate:
		g.gameOver = true
		g.gameResult = "Draw by stalemate"
	}
}

func (g *Game) updateMenu() {
	// Play vs AI
	if buttonClicked(100, 150, 200, 40) {
		g.playWithAI = true
	}

	// Local multiplayer
	if buttonClicked(320, 150, 200, 40) {
		g.playWithAI = false
	}

	// AI options
	if g.playWithAI {
		g.aiElo = slider(100, 230, 300, 800, 2000, g.aiElo)

		if buttonClicked(100, 270, 140, 40) {
			g.playerColor = chess.White
		}
		if buttonClicked(260, 270, 140, 40) {
			g.playerColor = chess.Black
		}
	}

	// Timer presets
	if buttonClicked(100, 340, 100, 40) {
		g.useTimer = false
	}
	if buttonClicked(210, 340, 100, 40) {
		g.useTimer = true
		g.timeSeconds = 180
	}
	if buttonClicked(320, 340, 100, 40) {
		g.useTimer = true
		g.timeSeconds = 300
	}

	// Start game
	if buttonClicked(180, 420, 200, 50) {
		g.startGame()
	}
	
}

func (g *Game) startGame() {
    g.chessGame = chess.NewGame()
    g.gameOver = false
    g.mode = ModePlaying

    if g.useTimer {
		g.clock = NewClock(g.timeSeconds)
		g.clock.last = time.Now()	
    }

    if g.playWithAI {
        g.engine.SetElo(g.aiElo)
        g.aiColor = oppositeColor(g.playerColor)
    } else {
        g.aiColor = chess.NoColor
    }
}


func oppositeColor(c chess.Color) chess.Color {
	if c == chess.White {
		return chess.Black
	}
	return chess.White
}

func (g *Game) Update() error {
	if g.mode == ModeMenu {
		g.updateMenu()
		return nil
	}

	mousePressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	
	if g.useTimer {
		g.clock.Update(g.chessGame.Position().Turn())
	}

	// restart the game with R
	if ebiten.IsKeyPressed(ebiten.KeyR) {
		g.resetGame()
		return nil
	}

	// Game over
	if g.gameOver {
		if ebiten.IsKeyPressed(ebiten.KeyR) {
			g.resetGame()
		}
		return nil
	}
	
	// Promotion Picker 
	if g.promotionFrom != nil {
		g.handlePromotion(mousePressed)
		g.mouseDown = mousePressed
		return nil
	}

	// AI turn
	if g.chessGame.Position().Turn() == g.aiColor {
		g.tryAIMove()
		g.mouseDown = mousePressed
		return nil
	}

	// Human input
	if mousePressed && !g.mouseDown {
		x, y := ebiten.CursorPosition()
		sq, ok := squareFromMouse(x, y)
		if ok {
			g.handleHumanClick(sq)
		}
	}

	// End-of-turn status check
	g.checkGameOver()
	g.mouseDown = mousePressed
	return nil
}

func findKingSquare(b *chess.Board, color chess.Color) (chess.Square, bool) {
	for sq := chess.Square(0); sq <= chess.H8; sq++ {
		p := b.Piece(sq)
		if p != chess.NoPiece &&
			p.Type() == chess.King &&
			p.Color() == color {
			return sq, true
		}
	}
	return chess.NoSquare, false
}

func kingInCheck(g *chess.Game) bool {
	moves := g.Moves()
	if len(moves) == 0 {
		return false
	}
	return moves[len(moves)-1].HasTag(chess.Check)
}

func (g *Game) drawPromotionPicker(screen *ebiten.Image) {
	if g.promotionFrom == nil || g.promotionTo == nil {
		return
	}

	file := int(g.promotionTo.File())
	rank := int(g.promotionTo.Rank())

	pieces := []chess.PieceType{
		chess.Queen,
		chess.Rook,
		chess.Bishop,
		chess.Knight,
	}

	for i, pt := range pieces {
		yRank := rank - i
		if g.chessGame.Position().Turn() == chess.Black {
			yRank = rank + i
		}

		x := float64(file) * tileSize
		y := float64(7-yRank) * tileSize

		// background
		ebitenutil.DrawRect(screen, x, y, tileSize, tileSize, color.RGBA{50, 50, 50, 230})

		piece := chess.NewPiece(pt, g.chessGame.Position().Turn())
		text.Draw(screen, piece.String(), pieceFace, int(x)+16, int(y)+128, color.White)

	}
}

func drawButton(screen *ebiten.Image, x, y, w, h int, label string, hovered bool) {
	bg := color.RGBA{80, 80, 80, 255}
	if hovered {
		bg = color.RGBA{120, 120, 120, 255}
	}

	ebitenutil.DrawRect(screen, float64(x), float64(y), float64(w), float64(h), bg)
	ebitenutil.DebugPrintAt(screen, label, x+10, y+10)
}

func buttonClicked(x, y, w, h int) bool {
	mx, my := ebiten.CursorPosition()
	if mx < x || mx > x+w || my < y || my > y+h {
		return false
	}
	return ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
}

func slider(x, y, w int, min, max, value int) int {
	mx, my := ebiten.CursorPosition()

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) &&
		my >= y-10 && my <= y+10 &&
		mx >= x && mx <= x+w {
		
		t := float64(mx-x) / float64(w)
		return min + int(t*float64(max-min))
	}
	return value
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, "Chess Puzzle Game", 160, 80)

	drawButton(screen, 100, 150, 200, 40, "Play vs AI", g.playWithAI)
	drawButton(screen, 320, 150, 200, 40, "Local Multiplayer", !g.playWithAI)

	if g.playWithAI {
		ebitenutil.DebugPrintAt(screen, "AI ELO", 100, 210)
		ebitenutil.DrawRect(screen, 100, 230, 300, 4, color.White)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", g.aiElo), 420, 220)

		drawButton(screen, 100, 270, 140, 40, "White", g.playerColor == chess.White)
		drawButton(screen, 260, 270, 140, 40, "Black", g.playerColor == chess.Black)
	}

	drawButton(screen, 100, 340, 100, 40, "No Timer", !g.useTimer)
	drawButton(screen, 210, 340, 100, 40, "3 min", g.useTimer && g.timeSeconds == 180)
	drawButton(screen, 320, 340, 100, 40, "5 min", g.useTimer && g.timeSeconds == 300)

	drawButton(screen, 180, 420, 200, 50, "START GAME", false)
}

func (g *Game) Draw(screen *ebiten.Image) {

	ebitenutil.DrawRect(screen, float64(boardSize), 0, panelWidth, boardSize, color.RGBA{30, 30, 30, 255})

	if g.mode == ModeMenu {
		g.drawMenu(screen)
		return
	}

	moves := g.chessGame.Moves()
	var lastMove *chess.Move
	if len(moves) > 0 {
		lastMove = moves[len(moves)-1]
	}
	lines := formatMoves(moves)

	y := 40
	for _, line := range lines {
		text.Draw(screen, line, textFace, boardSize+20, y, color.White)
		y += 32
	}

	// Draw chessboard
	for rank := chess.Rank8; rank >= chess.Rank1; rank-- {
		for file := chess.FileA; file <= chess.FileH; file++ {
			sq := chess.Square(int(file) + 8*int(rank))
			col := boardColors[(int(rank)+int(file))%2]
			ebitenutil.DrawRect(screen, float64(file)*tileSize, float64(7-rank)*tileSize, tileSize, tileSize, col)
			
			if g.selectedSquare != nil && sq == *g.selectedSquare {
				ebitenutil.DrawRect(
					screen,
					float64(file)*tileSize,
					float64(7-rank)*tileSize,
					tileSize,
					tileSize,
					color.RGBA{0, 255, 0, 80},
				)
			}
			// Highlight legal target squares
			if g.legalTargets != nil && g.legalTargets[sq] {
				if g.chessGame.Position().Board().Piece(sq) != chess.NoPiece {
        			// capture square
        			ebitenutil.DrawRect(
           			screen,
          			float64(file)*tileSize,
           			float64(7-rank)*tileSize,
            		tileSize,
            		tileSize,
            		color.RGBA{255, 0, 0, 80},
        			)
    			}
				ebitenutil.DrawRect(
					screen,
					float64(file)*tileSize,
					float64(7-rank)*tileSize,
					tileSize,
					tileSize,
					color.RGBA{0, 0, 0, 60},
				)

			}

			// Highlight last move
			if lastMove != nil && (sq == lastMove.S1() || sq == lastMove.S2()) {
				ebitenutil.DrawRect(screen, float64(file)*tileSize, float64(7-rank)*tileSize, tileSize, tileSize, color.RGBA{255, 255, 0, 80})
			}

			// Highlight King in Check 
			pos := g.chessGame.Position()
			if kingInCheck(g.chessGame) && pos.Status() != chess.Checkmate {
				if ks, ok := findKingSquare(pos.Board(), pos.Turn()); ok {
					file := int(ks.File())
					rank := 7 - int(ks.Rank())

					ebitenutil.DrawRect(screen,
						float64(file*tileSize),
						float64(rank*tileSize),
						float64(tileSize),
						float64(tileSize),
						color.RGBA{200, 0, 0, 120}, // translucent red
					)
				}
			}


			// Draw piece as string for now
			piece := g.chessGame.Position().Board().Piece(sq)
			if piece != chess.NoPiece {
				ebitenutil.DebugPrintAt(screen, piece.String(), int(file)*tileSize+20, int(7-rank)*tileSize+20)
			}
			text.Draw(screen, piece.String(), pieceFace, int(file)*tileSize+16, int(7-rank)*tileSize+128, color.Black)
		}
	}

	// Draw rank and file notations
	for rank := chess.Rank8; rank >= chess.Rank1; rank-- {
		colInv := boardColors[1-((int(rank)+int(chess.FileA))%2)]
		y := int(7-rank)*tileSize+40 
 	   	text.Draw(
    	    screen,
    	    rank.String(),
    	    textFace,
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
    	   	textFace,
     	 	int(file)*tileSize+116,
			1280, // bottom margin
      	 	colInv,
		)
	}

	g.drawPromotionPicker(screen)

	if g.useTimer {
    	g.clock.Draw(screen)	
	}

	if g.gameOver {
		ebitenutil.DrawRect(screen, 0, 0, float64(boardSize), float64(boardSize), color.RGBA{0, 0, 0, 180})
		text.Draw(screen, g.gameResult, textFace, boardSize/2-120, boardSize/2, color.White)
		text.Draw(screen, "Press R to Restart", textFace, boardSize/2-90, boardSize/2+40, color.White)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	engine, err := NewEngine("./engine/stockfish/stockfish-ubuntu-x86-64-bmi2")
	if err != nil {
		log.Fatal(err)
	}

	engine.SetElo(1200)

	game := &Game{
		chessGame: 	 chess.NewGame(),
		engine: 	 engine,
		aiColor: 	 chess.Black,
		aiElo: 		 1200,
		playWithAI:  true,
		mode: 		 ModeMenu,
		playerColor: chess.White,
		useTimer:	 false,
		timeSeconds: 300, // 5 min
	}
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Ebiten Chess")
	loadFonts()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
