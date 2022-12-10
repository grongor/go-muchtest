package match

import (
	"fmt"
)

func Len(len any) Matcher {
	return lenMatcher{len: len}
}

type lenMatcher struct {
	len any
}

func (m lenMatcher) Matches(actual any) (ok bool, desc string) {
	length, ok := m.getLength(actual)
	if !ok {
		return false, fmt.Sprintf("%s: len() panicked for value: %s", m.String(), formatValue(actual))
	}

	if ok, _ = ToMatcher(m.len).Matches(length); ok {
		return true, ""
	}

	return false, fmt.Sprintf("%s: got %d: %s", m.String(), length, formatValue(actual))
}

func (m lenMatcher) getLength(actual any) (length int, ok bool) {
	defer func() {
		if r := recover(); r != nil {
			length = 0
			ok = false
		}
	}()

	return reflectV(actual).Len(), true
}

func (m lenMatcher) String() string {
	return fmt.Sprintf("Len(%s)", formatValue(m.len))
}
