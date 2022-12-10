package match

import (
	"fmt"
	"reflect"
)

func Same(expected any) Matcher {
	return sameMatcher{same: expected}
}

func SamePointer(expected any) Matcher {
	return sameMatcher{same: expected, pointer: true}
}

type sameMatcher struct {
	same    any
	pointer bool
}

func (m sameMatcher) Matches(actual any) (ok bool, desc string) {
	if m.pointer {
		vSame := reflectV(m.same)
		vActual := reflectV(actual)

		if vSame.Kind() != reflect.Ptr || vActual.Kind() != reflect.Ptr {
			return false, fmt.Sprintf("%s: not pointer: %s", m.String(), formatValue(actual))
		}

		if vSame.Type() != vActual.Type() {
			return false, fmt.Sprintf("%s: not same type: %s", m.String(), formatValue(actual))
		}
	}

	defer func() {
		if recover() == nil && ok {
			return
		}

		ok = false

		switch reflectV(actual).Kind() {
		case reflect.Struct, reflect.Pointer:
			desc = fmt.Sprintf("%s: not same: %s%s", m.String(), formatValue(actual), getDiff(m.same, actual, nil))
		default:
			desc = fmt.Sprintf(
				"%s: not same: %T(%s)%s",
				m.String(), actual, formatValue(actual), getDiff(m.same, actual, nil),
			)
		}
	}()

	return m.same == actual, ""
}

func (m sameMatcher) String() string {
	return fmt.Sprintf("Same(%s)", formatValue(m.same))
}
