package cmd

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type Spinner interface {
	Start()
	Stop()
	Pause()
	Resume()
}

// SnakeSpinner renders a bouncing snake-like animation in the terminal.
type SnakeSpinner struct {
	width       int
	snakeLength int
	startColor  int
	tailColors  []int
	stopChan    chan struct{}
	pauseChan   chan struct{}
	resumeChan  chan struct{}
	running     bool
	mu          sync.Mutex
}

// NewSnakeSpinner creates a new instance of SnakeSpinner.
func NewSnakeSpinner(width, snakeLength, startColor int) *SnakeSpinner {
	return &SnakeSpinner{
		width:       width,
		snakeLength: snakeLength,
		startColor:  startColor,
		tailColors:  generateGradient(startColor, snakeLength-1),
		stopChan:    make(chan struct{}),
		pauseChan:   make(chan struct{}, 1),
		resumeChan:  make(chan struct{}, 1),
	}
}

func (b *SnakeSpinner) Start() {
	b.mu.Lock()
	if b.running {
		b.mu.Unlock()
		return
	}
	b.running = true
	b.mu.Unlock()

	go b.run()
}

func (b *SnakeSpinner) Stop() {
	close(b.stopChan)
}

func (b *SnakeSpinner) Pause() {
	select {
	case b.pauseChan <- struct{}{}:
	default:
	}
}

func (b *SnakeSpinner) Resume() {
	select {
	case b.resumeChan <- struct{}{}:
	default:
	}
}

func (b *SnakeSpinner) run() {
	const block = "█"
	const colorFormat = "\033[38;5;%dm%s\033[0m"
	snake := make([]int, b.snakeLength)
	for i := range snake {
		snake[i] = i
	}

	direction := 1
	left := 0
	right := b.width - 1
	paused := false

	for {
		select {
		case <-b.stopChan:
			fmt.Print("\r" + strings.Repeat(" ", b.width) + "\r")
			return
		case <-b.pauseChan:
			paused = true
		case <-b.resumeChan:
			paused = false
		default:
			if paused {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			line := make([]string, b.width)
			for i := range line {
				line[i] = " "
			}

			for idx, pos := range snake {
				if pos >= 0 && pos < b.width {
					if idx == b.snakeLength-1 {
						line[pos] = fmt.Sprintf(colorFormat, b.startColor, block)
					} else {
						color := b.tailColors[idx]
						line[pos] = fmt.Sprintf(colorFormat, color, block)
					}
				}
			}

			fmt.Printf("\r%s", strings.Join(line, ""))
			time.Sleep(150 * time.Millisecond)

			// Update snake position
			newHead := snake[b.snakeLength-1] + direction
			if newHead < left || newHead > right {
				direction = -direction
				newHead = snake[b.snakeLength-1] + direction
			}
			snake = append(snake[1:], newHead)
		}
	}
}

func generateGradient(startColor, steps int) []int {
	gradient := make([]int, steps)
	current := startColor
	decrement := 6
	for i := 0; i < steps; i++ {
		gradient[i] = current
		current -= decrement
		if current < 16 {
			current = 16
		}
	}
	// Reverse gradient
	for i, j := 0, steps-1; i < j; i, j = i+1, j-1 {
		gradient[i], gradient[j] = gradient[j], gradient[i]
	}
	return gradient
}
