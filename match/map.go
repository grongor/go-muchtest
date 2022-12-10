package match

import (
	"fmt"
	"strings"
)

func Map(keyAndValues ...any) Matcher {
	keyAndValuesCount := len(keyAndValues)
	if keyAndValuesCount%2 != 0 {
		return matcherErr(`Map(): keyAndValues must be pairs; eg.: "key", "lorem", "otherKey", 123`)
	}

	return mapMatcher{keyAndValues: keyAndValues}
}

func MapExact(keyAndValues ...any) Matcher {
	matcher := Map(keyAndValues...)
	if m, ok := matcher.(mapMatcher); ok {
		m.exact = true

		return m
	}

	return matcher
}

type mapMatcher struct {
	keyAndValues []any
	exact        bool
}

func (m mapMatcher) Matches(actual any) (ok bool, desc string) {
	keyAndValuesCount := len(m.keyAndValues)
	pairsCount := keyAndValuesCount / 2

	if m.exact {
		allMatcher := All(Map(m.keyAndValues...), Len(pairsCount))
		if ok, desc = allMatcher.Matches(actual); !ok {
			return false, m.String() + strings.TrimPrefix(desc, allMatcher.String())
		}

		return true, ""
	}

	if pairsCount == 1 {
		containsKeyValueMatcher := ContainsKeyValue(m.keyAndValues[0], m.keyAndValues[1])
		if ok, desc = containsKeyValueMatcher.Matches(actual); !ok {
			return false, m.String() + strings.TrimPrefix(desc, containsKeyValueMatcher.String())
		}

		return true, ""
	}

	matchers := make([]any, 0, pairsCount)

	for i := 1; i < keyAndValuesCount; i += 2 {
		matchers = append(matchers, ContainsKeyValue(m.keyAndValues[i-1], m.keyAndValues[i]))
	}

	allMatcher := All(matchers...)
	if ok, desc = allMatcher.Matches(actual); !ok {
		return false, m.String() + strings.TrimPrefix(desc, allMatcher.String())
	}

	return true, ""
}

func (m mapMatcher) String() string {
	args := strings.TrimSuffix(strings.TrimPrefix(formatValue(m.keyAndValues), "[]any{"), "}")

	if m.exact {
		return fmt.Sprintf("MapExact(%s)", args)
	}

	return fmt.Sprintf("Map(%s)", args)
}
