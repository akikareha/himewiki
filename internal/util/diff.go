package util

import (
	"github.com/pmezard/go-difflib/difflib"
)

func Diff(oldText, newText string) string {
	diff := difflib.UnifiedDiff{
		A: difflib.SplitLines(oldText),
		B: difflib.SplitLines(newText),
		FromFile: "old",
		ToFile:  "new",
		Context: 3,
	}
	text, _ := difflib.GetUnifiedDiffString(diff)
	return text
}
