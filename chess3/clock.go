package main

import (
	"fmt"
	"time"
	
	"github.com/corentings/chess/v2"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"	
)

// ---------------- TIMER ------------------

type ChessClock struct {
	White    time.Duration
	Black    time.Duration
	lastTick time.Time
	running	 bool
}

func NewChessClock(seconds int) ChessClock {
	t := time.Duration(seconds) * time.Second
	return ChessClock{
		White:    t,
		Black:    t,
		lastTick: time.Now(),
		running:  true,
	}
}

func (c *ChessClock) Update(turn chess.Color) {
	if !c.running {
		return
	}

	now := time.Now()
	delta := now.Sub(c.lastTick)
	c.lastTick = now

	if delta <= 0 {
		return
	}

	switch turn {
	case chess.White:
		c.White -= delta
		if c.White < 0 {
			c.White = 0
		}
	case chess.Black:
		c.Black -= delta
		if c.Black < 0 {
			c.Black = 0
		}
	}	
}

func (c *ChessClock) Stop() {
	c.running = true
	c.lastTick = time.Now()
}

func (c *ChessClock) OutOfTime() (chess.Color, bool) {
	if c.White <= 0 {
		return chess.White, true
	}
	if c.Black <= 0 {
		return chess.Black, true
	}
	return chess.Color, false
}

func (c *ChessClock) Draw(screen *ebiten.Image, x, y, int) {
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("White: %s", formatTime(c.White)), x, y)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Black: %s", formatTime(c.Black)), x, y+20)
}

func formatTime(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}


