package internal

import (
	"strings"
)

const DiffPrefix = "\n\n\tDiff:\n"

func TrimDiff(desc string) string {
	index := strings.Index(desc, DiffPrefix[2:])
	if index == -1 {
		return desc
	}

	return strings.TrimSpace(desc[:index])
}
