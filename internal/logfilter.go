package internal

import (
    "encoding/csv"
    "fmt"
    "strings"
)

// FilterLogs processes CSV formatted log data.
// It expects each CSV record to have at least three fields:
//   - Field 0: event type ("signal" or "exit")
//   - Field 1: code (signal number or exit code)
//   - Field 2: process name (or identifier)
// Example CSV row: signal,9,processName
func FilterLogs(input string) {
    // Create a new CSV reader from the input string.
    reader := csv.NewReader(strings.NewReader(input))
    records, err := reader.ReadAll()
    if err != nil {
        fmt.Println("Error reading CSV:", err)
        return
    }

    // Process each record.
    for _, record := range records {
        // Ensure we have at least three fields.
        if len(record) < 3 {
            fmt.Println("Incomplete record:", record)
            continue
        }
        eventType, code, procName := record[0], record[1], record[2]
        switch eventType {
        case "signal":
            fmt.Printf("%s terminated by signal %s\n", procName, code)
        case "exit":
            fmt.Printf("%s returned exit code %s\n", procName, code)
        default:
            // For any other event types, print the record as is.
            fmt.Println(record)
        }
    }
}

