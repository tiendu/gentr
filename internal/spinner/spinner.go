package spinner

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

type State struct {
	stop     chan struct{}
	pause    chan struct{}
	resume   chan struct{}
	running  bool
	mutex    sync.Mutex
	stopOnce sync.Once
}

func NewState() State {
	return State{
		stop:   make(chan struct{}),
		pause:  make(chan struct{}, 1),
		resume: make(chan struct{}, 1),
	}
}

func (s *State) Start(run func()) {
	s.mutex.Lock()
	if s.running {
		s.mutex.Unlock()
		return
	}
	s.running = true
	s.mutex.Unlock()
	go run()
}

func (s *State) Stop() {
	s.stopOnce.Do(func() { close(s.stop) })
}

func (s *State) Pause() {
	select {
	case s.pause <- struct{}{}:
	default:
	}
}

func (s *State) Resume() {
	select {
	case s.resume <- struct{}{}:
	default:
	}
}

type Snake struct {
	state       State
	width       int
	snakeLength int
	startColor  int
	tailColors  []int
	writer      io.Writer
	frameTime   time.Duration
}

func NewSnake(width, snakeLength, startColor int, writer io.Writer) *Snake {
	if writer == nil {
		writer = os.Stdout
	}
	if width < 1 {
		width = 1
	}
	if snakeLength < 1 {
		snakeLength = 1
	}
	if snakeLength > width {
		snakeLength = width
	}

	return &Snake{
		state:       NewState(),
		width:       width,
		snakeLength: snakeLength,
		startColor:  startColor,
		tailColors:  generateGradient(startColor, snakeLength-1),
		writer:      writer,
		frameTime:   150 * time.Millisecond,
	}
}

func (s *Snake) Start()  { s.state.Start(s.run) }
func (s *Snake) Stop()   { s.state.Stop() }
func (s *Snake) Pause()  { s.state.Pause() }
func (s *Snake) Resume() { s.state.Resume() }

func (s *Snake) colorForIndex(index int) int {
	if index == s.snakeLength-1 {
		return s.startColor
	}
	if index >= 0 && index < len(s.tailColors) {
		return s.tailColors[index]
	}
	return s.startColor
}

func (s *Snake) run() {
	const (
		block       = "█"
		colorFormat = "\x1b[38;5;%dm%s\x1b[0m"
	)

	snake := make([]int, s.snakeLength)
	for index := range snake {
		snake[index] = index
	}

	direction := 1
	paused := false

	for {
		select {
		case <-s.state.stop:
			fmt.Fprint(s.writer, "\r"+strings.Repeat(" ", s.width)+"\r")
			return
		case <-s.state.pause:
			paused = true
		case <-s.state.resume:
			paused = false
		default:
			if paused {
				time.Sleep(20 * time.Millisecond)
				continue
			}

			line := make([]string, s.width)
			for index := range line {
				line[index] = " "
			}
			for index, position := range snake {
				if position >= 0 && position < s.width {
					line[position] = fmt.Sprintf(colorFormat, s.colorForIndex(index), block)
				}
			}

			fmt.Fprintf(s.writer, "\r%s", strings.Join(line, ""))
			time.Sleep(s.frameTime)

			newHead := snake[s.snakeLength-1] + direction
			if newHead < 0 || newHead >= s.width {
				direction = -direction
				newHead = snake[s.snakeLength-1] + direction
			}
			snake = append(snake[1:], newHead)
		}
	}
}

func generateGradient(startColor, steps int) []int {
	if steps <= 0 {
		return nil
	}

	gradient := make([]int, steps)
	current := startColor
	for index := range gradient {
		gradient[index] = current
		current -= 6
		if current < 16 {
			current = 16
		}
	}
	for left, right := 0, len(gradient)-1; left < right; left, right = left+1, right-1 {
		gradient[left], gradient[right] = gradient[right], gradient[left]
	}
	return gradient
}
