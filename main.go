package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Sprite struct {
	Lines  []string
	Width  int
	Height int
}

type Entity struct {
	Sprite *Sprite
	X, Y   int
	DX, DY int
	Color  string // ANSI color code
}

type Cell struct {
	Rune  rune
	Color string
}

func getTerminalSize() (int, int) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return 24, 80 // Fallback
	}
	parts := strings.Fields(string(out))
	if len(parts) < 2 {
		return 24, 80
	}
	rows, _ := strconv.Atoi(parts[0])
	cols, _ := strconv.Atoi(parts[1])
	return rows, cols
}

func main() {
	rows, cols := getTerminalSize()

	// Hide cursor
	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h") // Show cursor on exit

	// Clear screen
	fmt.Print("\033[2J")

	// ANSI colors
	const (
		Reset   = "\033[0m"
		Green   = "\033[32m"
		Blue    = "\033[34m"
		Red     = "\033[31m"
		Yellow  = "\033[33m"
		Magenta = "\033[35m"
		Cyan    = "\033[36m"
	)

	fishColors := []string{Red, Yellow, Magenta, Cyan}

	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Helper to create random color for fish
	randFishColor := func() string {
		return fishColors[rand.Intn(len(fishColors))]
	}

	// Define sprites
	fishRight := Sprite{Lines: []string{"><>"}, Width: 3, Height: 1}
	fishLeft := Sprite{Lines: []string{"<><"}, Width: 3, Height: 1}
	largeFishRight := Sprite{Lines: []string{"><()))'>"}, Width: 8, Height: 1}
	largeFishLeft := Sprite{Lines: []string{"<'()))><"}, Width: 8, Height: 1}
	bubble := Sprite{Lines: []string{"o"}, Width: 1, Height: 1}
	plant := Sprite{Lines: []string{
		"  \\|/",
		" \\|/|",
		" |\\|/",
		" \\|/|",
	}, Width: 5, Height: 4}

	entities := []*Entity{
		{Sprite: &fishRight, X: 1, Y: 10, DX: 1, DY: 0, Color: randFishColor()},
		{Sprite: &fishLeft, X: cols - 4, Y: 15, DX: -1, DY: 0, Color: randFishColor()},
		{Sprite: &largeFishRight, X: 10, Y: 5, DX: 1, DY: 0, Color: randFishColor()},
		{Sprite: &largeFishLeft, X: cols - 10, Y: 8, DX: -1, DY: 0, Color: randFishColor()},
		{Sprite: &plant, X: 5, Y: rows - 4, DX: 0, DY: 0, Color: Green},
		{Sprite: &plant, X: cols - 10, Y: rows - 4, DX: 0, DY: 0, Color: Green},
	}

	// Create frame buffer
	buf := make([][]Cell, rows)
	for i := range buf {
		buf[i] = make([]Cell, cols)
	}

	for { // Run forever
		// Clear buffer
		for y := 0; y < rows; y++ {
			for x := 0; x < cols; x++ {
				buf[y][x] = Cell{Rune: ' ', Color: Reset}
			}
		}

		// Spawn bubbles randomly
		if rand.Float64() < 0.2 {
			entities = append(entities, &Entity{
				Sprite: &bubble,
				X:      rand.Intn(cols),
				Y:      rows - 2,
				DX:     0,
				DY:     -1,
				Color:  Blue,
			})
		}

		// Update and draw entities
		var newEntities []*Entity
		for _, e := range entities {
			e.X += e.DX
			e.Y += e.DY

			// Fish logic
			if e.Sprite == &fishLeft || e.Sprite == &fishRight || e.Sprite == &largeFishLeft || e.Sprite == &largeFishRight {
				if e.X < 0 || e.X > cols-e.Sprite.Width {
					e.DX = -e.DX
					e.X += e.DX * 2
					if e.DX > 0 {
						if e.Sprite == &fishLeft || e.Sprite == &fishRight {
							e.Sprite = &fishRight
						} else {
							e.Sprite = &largeFishRight
						}
					} else {
						if e.Sprite == &fishLeft || e.Sprite == &fishRight {
							e.Sprite = &fishLeft
						} else {
							e.Sprite = &largeFishLeft
						}
					}
				}
			}

			// Bubble logic
			if e.Sprite == &bubble {
				if e.Y < 0 {
					continue // Remove bubble that went off screen
				}
				// Wobble
				if rand.Float64() < 0.3 {
					e.X += rand.Intn(3) - 1
					if e.X < 0 {
						e.X = 0
					}
					if e.X >= cols {
						e.X = cols - 1
					}
				}
			}

			newEntities = append(newEntities, e)

			// Draw to buffer
			for sy, line := range e.Sprite.Lines {
				for sx, r := range line {
					py := e.Y + sy
					px := e.X + sx
					if py >= 0 && py < rows && px >= 0 && px < cols {
						buf[py][px] = Cell{Rune: r, Color: e.Color}
					}
				}
			}
		}
		entities = newEntities

		// Render buffer to screen
		fmt.Print("\033[H") // Move cursor to top-left
		for y := 0; y < rows; y++ {
			var sb strings.Builder
			lastColor := Reset
			for x := 0; x < cols; x++ {
				cell := buf[y][x]
				if cell.Color != lastColor {
					sb.WriteString(cell.Color)
					lastColor = cell.Color
				}
				sb.WriteRune(cell.Rune)
			}
			sb.WriteString(Reset) // Reset color at end of line
			fmt.Println(sb.String())
		}

		time.Sleep(100 * time.Millisecond)
	}
}
