package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
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

func generatePlant(height int) Sprite {
	lines := make([]string, height)
	patterns := []string{
		"  \\|/",
		" \\|/|",
		" |\\|/",
		" \\|/|",
	}
	for i := 0; i < height; i++ {
		lines[i] = patterns[i%len(patterns)]
	}
	return Sprite{Lines: lines, Width: 5, Height: height}
}

func main() {
	rows, cols := getTerminalSize()

	// Hide cursor
	fmt.Print("\033[?25l")

	// Handle Ctrl+C
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		// Restore terminal state
		fmt.Print("\033[?25h") // Show cursor
		r, _ := getTerminalSize()
		fmt.Printf("\033[%d;1H\n", r) // Move to bottom
		os.Exit(0)
	}()

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
	crab := Sprite{Lines: []string{"(/)o,,o(/)"}, Width: 10, Height: 1}

	minHeight := rows / 3
	maxHeight := rows / 2
	if minHeight < 1 {
		minHeight = 1
	}
	if maxHeight <= minHeight {
		maxHeight = minHeight + 1
	}

	plant1Height := minHeight + rand.Intn(maxHeight-minHeight+1)
	plant2Height := minHeight + rand.Intn(maxHeight-minHeight+1)

	plant1 := generatePlant(plant1Height)
	plant2 := generatePlant(plant2Height)

	entities := []*Entity{
		{Sprite: &fishRight, X: 1, Y: 10, DX: 1, DY: 0, Color: randFishColor()},
		{Sprite: &fishLeft, X: cols - 4, Y: 15, DX: -1, DY: 0, Color: randFishColor()},
		{Sprite: &largeFishRight, X: 10, Y: 5, DX: 1, DY: 0, Color: randFishColor()},
		{Sprite: &largeFishLeft, X: cols - 10, Y: 8, DX: -1, DY: 0, Color: randFishColor()},
		{Sprite: &plant1, X: 5, Y: rows - plant1Height, DX: 0, DY: 0, Color: Green},
		{Sprite: &plant2, X: cols - 10, Y: rows - plant2Height, DX: 0, DY: 0, Color: Green},
		{Sprite: &crab, X: cols / 2, Y: rows - 1, DX: 1, DY: 0, Color: Red},
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
			// Movement logic based on sprite
			if e.Sprite == &fishLeft || e.Sprite == &fishRight || e.Sprite == &largeFishLeft || e.Sprite == &largeFishRight {
				e.X += e.DX

				// Randomly change vertical direction
				if rand.Float64() < 0.05 {
					e.DY = rand.Intn(3) - 1 // -1, 0, or 1
				}
				e.Y += e.DY

				// Horizontal bounce
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

				// Vertical boundaries
				minY := 2
				maxY := rows - 6
				if e.Y < minY {
					e.Y = minY
					e.DY = 0
				}
				if e.Y > maxY {
					e.Y = maxY
					e.DY = 0
				}
			} else if e.Sprite == &crab {
				e.X += e.DX
				// Horizontal bounce for crab
				if e.X < 0 || e.X > cols-e.Sprite.Width {
					e.DX = -e.DX
					e.X += e.DX * 2
				}
			} else {
				// Bubble and Plant logic
				e.X += e.DX
				e.Y += e.DY
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
