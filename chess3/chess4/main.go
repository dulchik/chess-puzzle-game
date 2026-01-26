// NOTE: This is a STRUCTURAL refactor + design direction.
// It compiles conceptually, but you may need to adapt imports / small glue.
// Focus: OOP-style menu, timer placement, scrollable move list, move navigation.

package main

import (
    "fmt"
    "image/color"
    "time"

    "github.com/corentings/chess/v2"
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// ---------------- UI PRIMITIVES ----------------

type Button struct {
    X, Y, W, H int
    Label      string
    OnClick    func()
}

func (b *Button) Update() {
    mx, my := ebiten.CursorPosition()
    if mx >= b.X && mx <= b.X+b.W && my >= b.Y && my <= b.Y+b.H {
        if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
            b.OnClick()
        }
    }
}

func (b *Button) Draw(screen *ebiten.Image) {
    ebitenutil.DrawRect(screen, float64(b.X), float64(b.Y), float64(b.W), float64(b.H), color.RGBA{90, 90, 90, 255})
    ebitenutil.DebugPrintAt(screen, b.Label, b.X+10, b.Y+10)
}

// ---------------- TIMER ----------------

type ChessClock struct {
    White time.Duration
    Black time.Duration
    last  time.Time
    Active bool
}

func (c *ChessClock) Update(turn chess.Color) {
    if !c.Active { return }
    now := time.Now()
    delta := now.Sub(c.last)
    c.last = now

    if turn == chess.White {
        c.White -= delta
    } else {
        c.Black -= delta
    }
}

func (c *ChessClock) Draw(screen *ebiten.Image) {
    ebitenutil.DebugPrintAt(screen, fmt.Sprintf("White %s", formatTime(c.White)), 20, 20)
    ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Black %s", formatTime(c.Black)), 20, 40)
}

// ---------------- MOVE HISTORY (SCROLLABLE + NAV) ----------------

type MoveHistory struct {
    Moves      []*chess.Move
    Scroll     int
    ViewIndex  int // for arrow-key navigation
}

func (h *MoveHistory) Update() {
    if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
        h.Scroll = max(0, h.Scroll-1)
    }
    if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
        h.Scroll++
    }

    if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
        h.ViewIndex = max(0, h.ViewIndex-1)
    }
    if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
        h.ViewIndex = min(len(h.Moves), h.ViewIndex+1)
    }
}

func (h *MoveHistory) Draw(screen *ebiten.Image, x, y int) {
    start := h.Scroll
    end := min(len(h.Moves), start+12)

    for i := start; i < end; i++ {
        ebitenutil.DebugPrintAt(screen, h.Moves[i].String(), x, y+(i-start)*22)
    }
}

// ---------------- MENU (STATE-DRIVEN) ----------------

type StartMenu struct {
    Buttons []*Button
}

func NewStartMenu(g *Game) *StartMenu {
    return &StartMenu{Buttons: []*Button{
        {
            X: 100, Y: 150, W: 200, H: 40,
            Label: "Play vs AI",
            OnClick: func() { g.playWithAI = true },
        },
        {
            X: 320, Y: 150, W: 200, H: 40,
            Label: "Local Multiplayer",
            OnClick: func() { g.playWithAI = false },
        },
        {
            X: 180, Y: 420, W: 200, H: 50,
            Label: "START",
            OnClick: func() { g.startGame() },
        },
    }}
}

func (m *StartMenu) Update() {
    for _, b := range m.Buttons {
        b.Update()
    }
}

func (m *StartMenu) Draw(screen *ebiten.Image) {
    ebitenutil.DebugPrintAt(screen, "Chess Puzzle Game", 180, 80)
    for _, b := range m.Buttons {
        b.Draw(screen)
    }
}

// ---------------- GAME ----------------

type Game struct {
    chessGame *chess.Game

    mode int

    playWithAI bool

    clock   ChessClock
    history MoveHistory

    menu *StartMenu
}

const (
    ModeMenu = iota
    ModePlaying
)

func (g *Game) startGame() {
    g.chessGame = chess.NewGame()
    g.mode = ModePlaying
    g.clock.Active = true
    g.clock.last = time.Now()
}

func (g *Game) Update() error {
    switch g.mode {
    case ModeMenu:
        g.menu.Update()
        return nil

    case ModePlaying:
        g.clock.Update(g.chessGame.Position().Turn())
        g.history.Moves = g.chessGame.Moves()
        g.history.Update()
    }
    return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{20, 20, 20, 255})

	switch g.mode {
	case ModeMenu:
		g.menu.Draw(screen)
	case ModePlaying:
		g.clock.Draw(screen)
		g.history.Draw(screen, 1320, 60)
		ebitenutil.DebugPrintAt(screen, "GAME RUNNING", 20, 80)
	}

    if g.mode == ModeMenu {
        g.menu.Draw(screen)
        return
    }

    g.clock.Draw(screen)
    g.history.Draw(screen, 1320, 60)
}

func (g *Game) Layout(_, _ int) (int, int) {
    return 1680, 1280
}

// ---------------- UTILS ----------------

func formatTime(d time.Duration) string {
    if d < 0 { d = 0 }
    m := int(d.Minutes())
    s := int(d.Seconds()) % 60
    return fmt.Sprintf("%02d:%02d", m, s)
}

func min(a, b int) int { if a < b { return a }; return b }
func max(a, b int) int { if a > b { return a }; return b }


func main() {
	game := &Game{}
	game.menu = NewStartMenu(game)

	ebiten.SetWindowSize(1680, 1280)
	ebiten.SetWindowTitle("Chess Puzzle Game")

	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}

