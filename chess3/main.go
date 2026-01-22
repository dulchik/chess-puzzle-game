package main

import (
	"fmt"
	"image/color"
	"io/ioutil"
	"log"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"github.com/corentings/chess/v2"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
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

func (g *Game) resetGame() {
	g.chessGame = chess.NewGame()

	g.selectedSquare = nil
	g.legalTargets = nil

	g.promotionFrom = nil
	g.promotionTo = nil

	g.gameOver = false
}

func (g *Game) Update() error {
	if g.gameOver {
		if ebiten.IsKeyPressed(ebiten.KeyR) {
			g.resetGame()
		}
		return nil
	}

	mousePressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)

	if g.promotionFrom != nil {
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) && !g.mouseDown {
			g.promotionFrom = nil
			g.promotionTo = nil
			g.mouseDown = true
			return nil
		}

		if mousePressed && !g.mouseDown {
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
					g.mouseDown = mousePressed
					return nil
				}
			}
			g.promotionFrom = nil
			g.promotionTo = nil
		}
		
		g.mouseDown = mousePressed
		return nil
	}

	// Detect click (not hold)
	if mousePressed && !g.mouseDown {
		x, y := ebiten.CursorPosition()
		sq, ok := squareFromMouse(x, y)
		if ok {
			board := g.chessGame.Position().Board()
			piece := board.Piece(sq)

			// No piece selected yet -> try selecting
			if g.selectedSquare == nil {
				if piece != chess.NoPiece &&
					piece.Color() == g.chessGame.Position().Turn() {
					g.selectedSquare = &sq
					g.legalTargets = legalMovesFrom(g.chessGame.Position(), sq)
				}
			} else {
				// Attempt move
				if isPawnPromotion(g.chessGame.Position(), *g.selectedSquare, sq) {
					g.promotionFrom = g.selectedSquare
					g.promotionTo = &sq
					g.selectedSquare = nil
					g.legalTargets = nil
					g.mouseDown = mousePressed
					return nil
				}

				uci := g.selectedSquare.String() + sq.String()
				notation := chess.UCINotation{}
				move, err := notation.Decode(g.chessGame.Position(), uci)

				if err == nil {
					g.chessGame.Move(move, nil) 
					g.selectedSquare = nil
					g.legalTargets = nil
				} else {
					if piece != chess.NoPiece &&
						piece.Color() == g.chessGame.Position().Turn() {
						g.selectedSquare = &sq
						g.legalTargets = legalMovesFrom(g.chessGame.Position(), sq)
					} else {
						g.selectedSquare = nil
						g.legalTargets = nil
					}
				}
			}
		}
	}
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

	if ebiten.IsKeyPressed(ebiten.KeyR) {
		*g = Game{chessGame: chess.NewGame()}
	}
	

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

func (g *Game) Draw(screen *ebiten.Image) {

	ebitenutil.DrawRect(screen, float64(boardSize), 0, panelWidth, boardSize, color.RGBA{30, 30, 30, 255})

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
	game := &Game{chessGame: chess.NewGame()}
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Ebiten Chess")
	loadFonts()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
