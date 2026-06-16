package spinner

import (
	"bytes"
	"testing"
	"time"
)

func TestSnakeHeadUsesStartColor(t *testing.T) {
	spinner := NewSnake(30, 5, 81, &bytes.Buffer{})
	if got := spinner.colorForIndex(4); got != 81 {
		t.Fatalf("expected head color 81, got %d", got)
	}
}

func TestSnakeClampsInvalidDimensions(t *testing.T) {
	spinner := NewSnake(0, 10, 81, &bytes.Buffer{})
	if spinner.width != 1 || spinner.snakeLength != 1 {
		t.Fatalf("unexpected dimensions: width=%d length=%d", spinner.width, spinner.snakeLength)
	}
}

func TestSnakeStopIsIdempotent(t *testing.T) {
	spinner := NewSnake(5, 2, 81, &bytes.Buffer{})
	spinner.frameTime = time.Millisecond
	spinner.Start()
	time.Sleep(2 * time.Millisecond)
	spinner.Stop()
	spinner.Stop()
}

func TestGenerateGradient(t *testing.T) {
	if got := generateGradient(81, 0); got != nil {
		t.Fatalf("expected nil gradient, got %#v", got)
	}
	if got := generateGradient(81, 4); len(got) != 4 {
		t.Fatalf("expected four colors, got %#v", got)
	}
}
