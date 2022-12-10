package match

import (
	"fmt"
	"reflect"
	"strings"
)

func Fn(fn any) Matcher {
	const invalid = "Invalid Fn(): fn must be: func(actual any|typeOf(actual)) (ok bool[, desc string])"

	v := reflectV(fn)
	t := v.Type()
	argc := t.NumIn()
	outc := t.NumOut()
	tBool := reflectT(false)
	tString := reflectT("")

	if v.Kind() == reflect.Func && argc == 1 && outc == 1 || outc == 2 && t.Out(0) == tBool {
		if outc == 1 || t.Out(1) == tString {
			return fnMatcher{fn: func(value any) (ok bool, desc string) {
				vValue := reflectV(value)
				if vValue.Type().AssignableTo(t.In(0)) {
					result := v.Call([]reflect.Value{vValue})

					if result[0].Kind() == reflect.Bool {
						if outc == 2 && result[1].Kind() == reflect.String {
							return result[0].Interface().(bool), result[1].Interface().(string)
						} else if outc == 1 {
							return result[0].Interface().(bool), ""
						}
					}
				}

				return matcherErr(invalid, fn).Matches(nil)
			}}
		}
	}

	return matcherErr(invalid, fn)
}

type fnMatcher struct {
	fn func(any) (bool, string)
}

func (m fnMatcher) Matches(actual any) (ok bool, desc string) {
	if ok, desc = m.fn(actual); !ok {
		if desc == "" {
			return false, fmt.Sprintf("Fn(): not matched value: %s", formatValue(actual))
		}

		if strings.HasPrefix(desc, "Invalid Fn():") {
			return false, desc
		}

		return false, "Fn(): " + desc
	}

	return true, ""
}

func (m fnMatcher) String() string {
	return "Fn()"
}
