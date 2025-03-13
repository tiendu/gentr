package cmd

import (
    "fmt"
    "strings"
    "time"
)

// generateGradient produces a slice of ANSI 256 color codes forming a gradient.
func generateGradient(startColor int, steps int) []int {
    gradient := make([]int, steps)
    current := startColor
    decrement := 6
    for i := 0; i < steps; i++ {
        gradient[i] = current
        current -= decrement
        if current < 16 { // Avoid going too dark.
            current = 16
        }
    }
    // Reverse the gradient so that the tail closest to the head is brightest.
    for i, j := 0, len(gradient)-1; i < j; i, j = i+1, j-1 {
        gradient[i], gradient[j] = gradient[j], gradient[i]
    }
    return gradient
}

// BounceSpinner displays a bouncing snake animation with a fading tail using gradient colors.
func BounceSpinner(done chan struct{}, control chan string) {
    width := 30         // Total width of the output line.
    snakeLength := 5    // Number of blocks in the snake.

    // ANSI color format string.
    const colorFormat = "\033[38;5;%dm%s\033[0m"

    // Initialize snake positions: starting at the left.
    snake := make([]int, snakeLength)
    for i := 0; i < snakeLength; i++ {
        snake[i] = i
    }

    direction := 1  // 1 means moving right; -1 means moving left.
    leftBoundary := 0
    rightBoundary := width - 1

    paused := false

    // Define the starting color for the tail gradient.
    startColor := 81  // A shade in the blue/cyan range.
    // Generate gradient colors for the tail segments (snakeLength-1 segments).
    tailColors := generateGradient(startColor, snakeLength-1)

    for {
        select {
        case <-done:
            // Clear the line before exiting.
            fmt.Print("\r" + strings.Repeat(" ", width) + "\r")
            return
        case cmd := <-control:
            // Handle pause/resume commands.
            if cmd == "pause" {
                paused = true
            } else if cmd == "resume" {
                paused = false
            }
        default:
            if paused {
                time.Sleep(100 * time.Millisecond)
                continue
            }
            // Create a slice for the current frame.
            line := make([]string, width)
            for i := 0; i < width; i++ {
                line[i] = " " // fill with spaces.
            }
            // Draw the snake.
            for idx, pos := range snake {
                if pos >= 0 && pos < width {
                    block := "█"
                    if idx == snakeLength-1 {
                        // The head: use starting color.
                        line[pos] = fmt.Sprintf(colorFormat, startColor, block)
                    } else {
                        // The tail: use gradient colors.
                        colorIdx := idx
                        if colorIdx >= len(tailColors) {
                            colorIdx = len(tailColors) - 1
                        }
                        line[pos] = fmt.Sprintf(colorFormat, tailColors[colorIdx], block)
                    }
                }
            }
            // Print the current frame.
            fmt.Printf("\r%s", strings.Join(line, ""))
            time.Sleep(150 * time.Millisecond)

            // Update snake position: compute new head position.
            newHead := snake[snakeLength-1] + direction
            if newHead < leftBoundary || newHead > rightBoundary {
                direction = -direction
                newHead = snake[snakeLength-1] + direction
            }
            // Shift the snake: remove the first element and append the new head.
            snake = append(snake[1:], newHead)
        }
    }
}

