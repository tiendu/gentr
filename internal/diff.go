package internal

import (
   "fmt"
   "math"
   "strings"
)

// DiffLines computes a diff between oldLines and newLines using an LCS algorithm.
// It returns two slices: added lines and removed lines.
func DiffLines(oldLines, newLines []string) (added []string, removed []string) {
   m, n := len(oldLines), len(newLines)
   // Create a DP table for LCS.
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

   // Backtrack to get the diff.
   var diff []string
   i, j := m, n
   for i > 0 && j > 0 {
      if oldLines[i-1] == newLines[j-1] {
         diff = append([]string{" " + oldLines[i-1]}, diff...)
         i--
         j--
      } else if dp[i-1][j] >= dp[i][j-1] {
         diff = append([]string{"- " + oldLines[i-1]}, diff...)
         i--
      } else {
         diff = append([]string{"+ " + newLines[j-1]}, diff...)
         j--
      }
   }
   for i > 0 {
      diff = append([]string{"- " + oldLines[i-1]}, diff...)
      i--
   }
   for j > 0 {
      diff = append([]string{"+ " + newLines[j-1]}, diff...)
      j--
   }

   // Separate added and removed lines.
   for _, line := range diff {
      if strings.HasPrefix(line, "+ ") {
         added = append(added, strings.TrimPrefix(line, "+ "))
      } else if strings.HasPrefix(line, "- ") {
         removed = append(removed, strings.TrimPrefix(line, "- "))
      }
   }
   return added, removed
}

