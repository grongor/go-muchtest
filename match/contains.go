package match

import (
	"fmt"
	"reflect"
	"strings"
)

func Contains(value any) Matcher {
	return newContainsMatcher("Contains", nil, value)
}

func ContainsKey(key any) Matcher {
	return newContainsMatcher("ContainsKey", key, nil)
}

func ContainsKeyValue(key, value any) Matcher {
	return newContainsMatcher("ContainsKeyValue", key, value)
}

func Prefix(prefix any) Matcher {
	m := newContainsMatcher("Prefix", nil, prefix)
	m.prefix = true

	return m
}

func Suffix(suffix any) Matcher {
	m := newContainsMatcher("Suffix", nil, suffix)
	m.suffix = true

	return m
}

func newContainsMatcher(name string, key, value any) containsMatcher {
	return containsMatcher{name: name, key: key, value: value}
}

type containsMatcher struct {
	key, value     any
	name           string
	prefix, suffix bool
}

func (m containsMatcher) Matches(actual any) (ok bool, desc string) {
	vContainer := reflectV(actual)
	tContainer := vContainer.Type()

	switch kind := tContainer.Kind(); kind {
	case reflect.String:
		return m.stringContains(vContainer)
	case reflect.Slice:
		return m.sliceContains(vContainer)
	case reflect.Map:
		return m.mapContains(vContainer)
	case reflect.Struct:
		return m.structContains(vContainer)
	case reflect.Pointer:
		return m.Matches(vContainer.Elem())
	default:
		return false, fmt.Sprintf("%s: unsupported container kind: %s", m.String(), kind)
	}
}

func (m containsMatcher) String() string {
	switch true {
	case m.key == nil:
		return fmt.Sprintf("%s(%s)", m.name, formatValue(m.value))
	case m.value == nil:
		return fmt.Sprintf("%s(%s)", m.name, formatValue(m.key))
	default:
		return fmt.Sprintf("%s(%s, %s)", m.name, formatValue(m.key), formatValue(m.value))
	}
}

func (m containsMatcher) stringContains(container reflect.Value) (bool, string) {
	if m.key != nil {
		return false, m.String() + ": can't be use when typeOf(actual) == string"
	}

	if _, ok := m.value.(Matcher); ok {
		return false, m.String() + ": value can't be a Matcher when typeOf(actual) == string"
	}

	expected, ok := toString(m.value)
	if !ok {
		return false, fmt.Sprintf("%s: invalid value; possible types: %s", m.String(), toStringTypes())
	}

	if m.prefix {
		if strings.HasPrefix(container.String(), expected) {
			return true, ""
		}

		return false, fmt.Sprintf(`%s: not prefixed: %q`, m.String(), container)
	}

	if m.suffix {
		if strings.HasSuffix(container.String(), expected) {
			return true, ""
		}

		return false, fmt.Sprintf(`%s: not suffixed: %q`, m.String(), container)
	}

	if strings.Contains(container.String(), expected) {
		return true, ""
	}

	return false, fmt.Sprintf(`%s: not contained in: %q`, m.String(), container)
}

func (m containsMatcher) sliceContains(container reflect.Value) (bool, string) {
	withValue := m.value != nil

	if m.key != nil {
		i, ok := m.key.(int)
		if !ok {
			return false, m.String() + ": key must be int when typeOf(actual) == []any"
		}

		if i >= container.Len() {
			return false, fmt.Sprintf("%s: key out of range: len(actual) == %d", m.String(), container.Len())
		}

		value := container.Index(i).Interface()

		if withValue {
			if ok, _ = ToMatcher(m.value).Matches(value); ok {
				return true, ""
			}

			return false, fmt.Sprintf("%s: value not matched: %s", m.String(), formatValue(value))
		}

		if container.Index(i).IsZero() {
			return false, fmt.Sprintf("%s: value is zero: %v", m.String(), value)
		}

		if m.value == nil {
			return true, ""
		}
	}

	if m.prefix {
		return false, "not implemented"
	}

	if m.suffix {
		return false, "not implemented"
	}

	matcher := ToMatcher(m.value)

	for i := 0; i < container.Len(); i++ {
		if ok, _ := matcher.Matches(container.Index(i).Interface()); ok {
			return true, ""
		}
	}

	return false, fmt.Sprintf(`%s: no such value in: %s`, m.String(), formatValue(container))
}

func (m containsMatcher) mapContains(vContainer reflect.Value) (bool, string) {
	if m.suffix || m.prefix {
		return false, m.String() + ": can't be use when typeOf(actual) == map"
	}

	withValue := m.value != nil
	mapKeys := vContainer.MapKeys()

	if m.key != nil {
		for i := 0; i < len(mapKeys); i++ {
			vIndex := mapKeys[i]
			index := vIndex.Interface()
			if ok, _ := ToMatcher(m.key).Matches(index); !ok {
				continue
			}

			if !withValue {
				return true, ""
			}

			value := vContainer.MapIndex(vIndex).Interface()
			if ok, _ := ToMatcher(m.value).Matches(value); ok {
				return true, ""
			}

			return false, fmt.Sprintf("%s: value not matched: %s", m.String(), formatValue(value))
		}

		return false, fmt.Sprintf("%s: no such key in: %s", m.String(), formatValue(vContainer))
	}

	for i := 0; i < len(mapKeys); i++ {
		if ok, _ := ToMatcher(m.value).Matches(vContainer.MapIndex(mapKeys[i]).Interface()); ok {
			return true, ""
		}
	}

	return false, fmt.Sprintf(`%s: no such value in: %s`, m.String(), formatValue(vContainer))
}

func (m containsMatcher) structContains(vContainer reflect.Value) (bool, string) {
	if m.suffix || m.prefix {
		return false, m.String() + ": can't be use when typeOf(actual) == struct"
	}

	fieldsCount := vContainer.NumField()

	if m.key != nil {
		tContainer := vContainer.Type()
		for i := 0; i < fieldsCount; i++ {
			vField := vContainer.Field(i)
			if !vField.CanInterface() {
				continue
			}

			field := tContainer.Field(i).Name
			if ok, _ := ToMatcher(m.key).Matches(field); !ok {
				continue
			}

			if m.value == nil {
				return true, ""
			}

			value := vField.Interface()
			if ok, _ := ToMatcher(m.value).Matches(value); ok {
				return true, ""
			}

			return false, fmt.Sprintf("%s: value not matched: %s", m.String(), formatValue(value))
		}

		return false, fmt.Sprintf("%s: no such field in: %s", m.String(), formatValue(vContainer))
	}

	for i := 0; i < fieldsCount; i++ {
		vField := vContainer.Field(i)
		if !vField.CanInterface() {
			continue
		}

		if ok, _ := ToMatcher(m.value).Matches(vField.Interface()); ok {
			return true, ""
		}
	}

	return false, fmt.Sprintf(`%s: no such value in: %s`, m.String(), formatValue(vContainer))
}
