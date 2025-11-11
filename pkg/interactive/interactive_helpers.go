package interactive

import (
	"strings"

	"github.com/turbot/go-kit/helpers"
)

type queryCompletionInfo struct {
	Table        string
	EditingTable bool
}

func getQueryInfo(text string) *queryCompletionInfo {
	table := getTable(text)
	prevWord := getPreviousWord(text)

	return &queryCompletionInfo{
		Table:        table,
		EditingTable: isEditingTable(text, prevWord),
	}
}

func isEditingTable(text string, prevWord string) bool {
	// We're editing a table if:
	// 1. The previous word is "from"
	// 2. AND the text ends with a space (meaning cursor is after "from " ready for input)
	// If text doesn't end with space, we're already typing the table name
	return prevWord == "from" && strings.HasSuffix(text, " ")
}

func getTable(text string) string {
	// split on space and remove empty results - they occur if there is a double space
	split := helpers.RemoveFromStringSlice(strings.Split(text, " "), "")

	for idx, word := range split {
		if word == "from" {
			if idx+1 < len(split) {
				return split[idx+1]
			}
		}
	}
	return ""
}

func getPreviousWord(text string) string {
	// If text ends with space(s), we want the word immediately before those spaces
	// e.g., "from " should return "from"
	// e.g., "select * from " should return "from"

	// First, find the last space
	finalSpace := strings.LastIndex(text, " ")
	if finalSpace == -1 {
		return ""
	}

	// Find the last non-space character before that final space
	lastNotSpace := lastIndexByteNot(text[:finalSpace], ' ')
	if lastNotSpace == -1 {
		// No non-space character before the final space means text is all spaces
		return ""
	}

	// Find the space before the last word (if any)
	prevSpace := strings.LastIndex(text[:lastNotSpace+1], " ")
	if prevSpace == -1 {
		// No space before means the last word starts at the beginning
		// Return from start to lastNotSpace
		return text[0 : lastNotSpace+1]
	}

	// Return the word between prevSpace and lastNotSpace
	return text[prevSpace+1 : lastNotSpace+1]
}

func lastIndexByteNot(s string, c byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] != c {
			return i
		}
	}
	return -1
}

// if there are no spaces this is the first word
func isFirstWord(text string) bool {
	return strings.LastIndex(text, " ") == -1
}

// split the string by spaces and return the last segment
func lastWord(text string) string {
	return text[strings.LastIndex(text, " "):]
}

//
// keeping this around because we may need
// to revisit exit on non-darwin platforms.
// as per line #128
//
//
// https://github.com/c-bata/go-prompt/issues/59
// func exit(_ *prompt.Buffer) {
// 	fmt.Println("Ctrl+D :: exitCallback")
// 	panic(utils.ExitCode(0))
// }
