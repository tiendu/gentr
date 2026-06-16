package diff

type Kind string

const (
	Added    Kind = "ADD"
	Removed  Kind = "REM"
	Modified Kind = "MOD"
)

type Change struct {
	LineNumber int
	Kind       Kind
	Text       string
}

func Lines(oldLines, newLines []string) []Change {
	oldLength, newLength := len(oldLines), len(newLines)
	table := make([][]int, oldLength+1)
	for index := range table {
		table[index] = make([]int, newLength+1)
	}

	for oldIndex := 1; oldIndex <= oldLength; oldIndex++ {
		for newIndex := 1; newIndex <= newLength; newIndex++ {
			if oldLines[oldIndex-1] == newLines[newIndex-1] {
				table[oldIndex][newIndex] = table[oldIndex-1][newIndex-1] + 1
			} else if table[oldIndex-1][newIndex] >= table[oldIndex][newIndex-1] {
				table[oldIndex][newIndex] = table[oldIndex-1][newIndex]
			} else {
				table[oldIndex][newIndex] = table[oldIndex][newIndex-1]
			}
		}
	}

	changes := make([]Change, 0)
	oldIndex, newIndex := oldLength, newLength
	for oldIndex > 0 && newIndex > 0 {
		switch {
		case oldLines[oldIndex-1] == newLines[newIndex-1]:
			oldIndex--
			newIndex--
		case table[oldIndex-1][newIndex] >= table[oldIndex][newIndex-1]:
			changes = prepend(changes, Change{LineNumber: oldIndex, Kind: Removed, Text: oldLines[oldIndex-1]})
			oldIndex--
		default:
			changes = prepend(changes, Change{LineNumber: newIndex, Kind: Added, Text: newLines[newIndex-1]})
			newIndex--
		}
	}

	for oldIndex > 0 {
		changes = prepend(changes, Change{LineNumber: oldIndex, Kind: Removed, Text: oldLines[oldIndex-1]})
		oldIndex--
	}
	for newIndex > 0 {
		changes = prepend(changes, Change{LineNumber: newIndex, Kind: Added, Text: newLines[newIndex-1]})
		newIndex--
	}

	return changes
}

func CombineModifications(changes []Change) []Change {
	combined := make([]Change, 0, len(changes))
	for index := 0; index < len(changes); index++ {
		if index+1 < len(changes) &&
			changes[index].Kind == Added &&
			changes[index+1].Kind == Removed &&
			changes[index].LineNumber == changes[index+1].LineNumber {
			combined = append(combined, Change{
				LineNumber: changes[index].LineNumber,
				Kind:       Modified,
				Text:       changes[index+1].Text + " -> " + changes[index].Text,
			})
			index++
			continue
		}
		combined = append(combined, changes[index])
	}
	return combined
}

func prepend(changes []Change, change Change) []Change {
	changes = append(changes, Change{})
	copy(changes[1:], changes)
	changes[0] = change
	return changes
}
