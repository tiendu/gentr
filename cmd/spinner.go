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

type SpinnerState struct {
	stopChan    chan struct{}
	pauseChan   chan struct{}
	resumeChan  chan struct{}
	running     bool
	mu          sync.Mutex
}

func NewSpinnerState() SpinnerState {
	return SpinnerState{
		stopChan:    make(chan struct{}),
		pauseChan:   make(chan struct{}, 1),
		resumeChan:  make(chan struct{}, 1),
	}
}

func (s *SpinnerState) Start(run func()) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()
	go run()
}

func (s *SpinnerState) Stop() {
	close(s.stopChan)
}

func (s *SpinnerState) Pause() {
	select {
	case s.pauseChan <- struct{}{}:
	default:
	}
}

func (s *SpinnerState) Resume() {
	select {
	case s.resumeChan <- struct{}{}:
	default:
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

// =============================
// SnakeSpinner Implementation
// =============================

type SnakeSpinner struct {
	state       SpinnerState
	width       int
	snakeLength int
	startColor  int
	tailColors  []int
}

func NewSnakeSpinner(width, snakeLength, startColor int) *SnakeSpinner {
	return &SnakeSpinner{
		state:       NewSpinnerState(),
		width:       width,
		snakeLength: snakeLength,
		startColor:  startColor,
		tailColors:  generateGradient(startColor, snakeLength-1),
	}
}

func (s *SnakeSpinner) Start()  { s.state.Start(s.run) }
func (s *SnakeSpinner) Stop()   { s.state.Stop() }
func (s *SnakeSpinner) Pause()  { s.state.Pause() }
func (s *SnakeSpinner) Resume() { s.state.Resume() }

func (s *SnakeSpinner) run() {
	const block = "█"
	const colorFormat = "\033[38;5;%dm%s\033[0m"
	snake := make([]int, s.snakeLength)
	for i := range snake {
		snake[i] = i
	}

	direction := 1
	left := 0
	right := s.width - 1
	paused := false

	for {
		select {
		case <-s.state.stopChan:
			fmt.Print("\r" + strings.Repeat(" ", s.width) + "\r")
			return
		case <-s.state.pauseChan:
			paused = true
		case <-s.state.resumeChan:
			paused = false
		default:
			if paused {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			line := make([]string, s.width)
			for i := range line {
				line[i] = " "
			}

			for idx, pos := range snake {
				if pos >= 0 && pos < s.width {
					color := s.tailColors[idx]
					if idx == s.snakeLength-1 {
						color = s.startColor
					}
					line[pos] = fmt.Sprintf(colorFormat, color, block)
				}
			}

			fmt.Printf("\r%s", strings.Join(line, ""))
			time.Sleep(150 * time.Millisecond)

			newHead := snake[s.snakeLength-1] + direction
			if newHead < left || newHead > right {
				direction = -direction
				newHead = snake[s.snakeLength-1] + direction
			}
			snake = append(snake[1:], newHead)
		}
	}
}

// =============================
// DotLineSpinner Implementation
// =============================

type DotLineSpinner struct {
	state      SpinnerState
	frameTime  time.Duration
	startColor int
	length     int
	frames     []string
}

func NewDotLineSpinner(frameTime time.Duration, startColor int, length int, frames []string) *DotLineSpinner {
	return &DotLineSpinner{
		state:      NewSpinnerState(),
		frameTime:  frameTime,
		startColor: startColor,
		length:     length,
		frames:     frames,
	}
}

func (s *DotLineSpinner) Start()  { s.state.Start(s.run) }
func (s *DotLineSpinner) Stop()   { s.state.Stop() }
func (s *DotLineSpinner) Pause()  { s.state.Pause() }
func (s *DotLineSpinner) Resume() { s.state.Resume() }

func (s *DotLineSpinner) run() {
	const colorFormat = "\033[38;5;%dm%s\033[0m"
	if len(s.frames) == 0 {
		s.frames = []string{"·", "•", "◦", "○", "◎", "◉", "●", "◉", "◎", "○", "◦", "•", "·"}
	}
	gradient := generateGradient(s.startColor, s.length)

	step := 0
	paused := false

	for {
		select {
		case <-s.state.stopChan:
			fmt.Print("\r" + strings.Repeat(" ", s.length*4) + "\r")
			return
		case <-s.state.pauseChan:
			paused = true
		case <-s.state.resumeChan:
			paused = false
		default:
			if paused {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			var output []string
			for i := 0; i < s.length; i++ {
				ch := s.frames[(step+i)%len(s.frames)]
				color := gradient[i]
				output = append(output, fmt.Sprintf(colorFormat, color, ch))
			}
			fmt.Printf("\r%s", strings.Join(output, " "))
			time.Sleep(s.frameTime)
			step++
		}
	}
}

// ================================
// RollingShapeSpinner Implementation
// ================================

type RollingShapeSpinner struct {
	state      SpinnerState
	frameTime  time.Duration
	startColor int
	length     int
	shapes     []string
}

func NewRollingShapeSpinner(frameTime time.Duration, startColor int, length int, shapes []string) *RollingShapeSpinner {
	return &RollingShapeSpinner{
		state:      NewSpinnerState(),
		frameTime:  frameTime,
		startColor: startColor,
		length:     length,
		shapes:     shapes,
	}
}

func (s *RollingShapeSpinner) Start()  { s.state.Start(s.run) }
func (s *RollingShapeSpinner) Stop()   { s.state.Stop() }
func (s *RollingShapeSpinner) Pause()  { s.state.Pause() }
func (s *RollingShapeSpinner) Resume() { s.state.Resume() }

func (s *RollingShapeSpinner) run() {
	const colorFormat = "\033[38;5;%dm%s\033[0m"
	if len(s.shapes) == 0 {
		// s.shapes = []string{"→", "↘", "↓", "↙", "←", "↖", "↑", "↗"}
		// s.shapes = []string{"░", "▒", "▓", "█", "▓", "▒", "░"}
		s.shapes = []string{"▖", "▘", "▝", "▗"}
		// s.shapes = []string{"◢", "◣", "◤", "◥"}
		// s.shapes = []string{"◇", "◆", "◈", "◇"}
		// s.shapes = []string{"⠁", "⠂", "⠄", "⡀", "⢀", "⠠", "⠐", "⠈"}
	}
	gradient := generateGradient(s.startColor, s.length)

	step := 0
	paused := false

	for {
		select {
		case <-s.state.stopChan:
			fmt.Print("\r" + strings.Repeat(" ", s.length*2) + "\r")
			return
		case <-s.state.pauseChan:
			paused = true
		case <-s.state.resumeChan:
			paused = false
		default:
			if paused {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			output := make([]string, s.length)
			for i := range output {
				output[i] = " "
			}

			pos := s.length - 1 - (step % s.length)                // move right to left
			shape := s.shapes[step%len(s.shapes)]                 // rotating shape
			output[pos] = fmt.Sprintf(colorFormat, gradient[pos], shape)

			fmt.Printf("\r%s", strings.Join(output, " "))
			time.Sleep(s.frameTime)
			step++
		}
	}
}
