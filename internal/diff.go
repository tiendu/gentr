package internal

import (
    "math"

    "gentr/internal/utils"
)

type DiffChange struct {
    LineNumber int
    Type       string // "ADD" or "REM"
    Text       string
}

// DiffLines computes a diff between oldLines and newLines using an LCS algorithm.
// It returns a slice of DiffChange items representing added or removed lines.
// For additions, the line number is taken from newLines, and for removals, from oldLines.
func DiffLines(oldLines, newLines []string) []DiffChange {
    m, n := len(oldLines), len(newLines)
    // Create DP table for LCS.
    dp := make([][]int, m+1)
    for i := range dp {
        dp[i] = make([]int, n+1)
    }
    for i := 1; i <= m; i++ {
        for j := 1; j <= n; j++ {
            if oldLines[i-1] == newLines[j-1] {
                dp[i][j] = dp[i-1][j-1] + 1
            } else {
                dp[i][j] = int(math.Max(float64(dp[i-1][j]), float64(dp[i][j-1])))
            }
        }
    }

    var changes []DiffChange
    i, j := m, n
    // Backtrack through the DP table to determine diff.
    for i > 0 && j > 0 {
        if oldLines[i-1] == newLines[j-1] {
            i--
            j--
        } else if dp[i-1][j] >= dp[i][j-1] {
            changes = append([]DiffChange{{
                LineNumber: i,
                Type:       "REM",
                Text:       utils.TruncateLine(oldLines[i-1], 60),
            }}, changes...)
            i--
        } else {
            changes = append([]DiffChange{{
                LineNumber: j,
                Type:       "ADD",
                Text:       utils.TruncateLine(newLines[j-1], 60),
            }}, changes...)
            j--
        }
    }
    // Process remaining lines if any.
    for i > 0 {
        changes = append([]DiffChange{{
            LineNumber: i,
            Type:       "REM",
            Text:       utils.TruncateLine(oldLines[i-1], 60),
        }}, changes...)
        i--
    }
    for j > 0 {
        changes = append([]DiffChange{{
            LineNumber: j,
            Type:       "ADD",
            Text:       utils.TruncateLine(newLines[j-1], 60),
        }}, changes...)
        j--
    }
    return changes
}

