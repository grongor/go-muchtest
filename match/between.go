package match

import (
	"fmt"
	"math/big"
	"reflect"
)

func Between(min, max any) Matcher {
	return betweenMatcher{min: min, max: max}
}

func BetweenStrict(min, max any) Matcher {
	return betweenMatcher{min: min, max: max, strict: true}
}

type betweenMatcher struct {
	min, max any
	strict   bool
}

func (m betweenMatcher) Matches(actual any) (ok bool, desc string) {
	vMin := reflectV(m.min)
	vMax := reflectV(m.max)
	vActual := reflectV(actual)

	if m.strict {
		if vMin.Type() != vMax.Type() {
			return false, fmt.Sprintf("%s: min and max have different types: min(%T) max(%T)", m.String(), m.min, m.max)
		}

		if vMin.Type() != vActual.Type() {
			return false, fmt.Sprintf("%s: different type: %T(%s)", m.String(), actual, formatValue(actual))
		}
	}

	var values [3]any
	var withFloat, withFloat32 bool

	for i, v := range [...]reflect.Value{vMin, vMax, vActual} {
		switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			values[i] = big.NewInt(v.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			values[i] = big.NewInt(0).SetUint64(v.Uint())
		case reflect.Float32:
			withFloat32 = true

			fallthrough
		case reflect.Float64:
			values[i] = big.NewFloat(v.Float())
			withFloat = true
		default:
			return false, fmt.Sprintf("%s: not comparable: %T(%s)", m.String(), actual, formatValue(actual))
		}
	}

	if withFloat {
		for i, v := range values {
			if intValue, ok := v.(*big.Int); ok {
				values[i] = big.NewFloat(0).SetInt(intValue)
			} else if withFloat32 {
				f32, _ := values[i].(*big.Float).Float32()
				values[i] = big.NewFloat(float64(f32))
			}
		}

		min, max, value := values[0].(*big.Float), values[1].(*big.Float), values[2].(*big.Float)

		if value.Cmp(min) >= 0 && value.Cmp(max) <= 0 {
			return true, ""
		}
	} else {
		min, max, value := values[0].(*big.Int), values[1].(*big.Int), values[2].(*big.Int)

		if value.Cmp(min) >= 0 && value.Cmp(max) <= 0 {
			return true, ""
		}
	}

	return false, fmt.Sprintf("%s: outside range: %s", m.String(), formatValue(actual))
}

func (m betweenMatcher) String() string {
	if m.strict {
		return fmt.Sprintf("BetweenStrict(%s, %s)", formatValue(m.min), formatValue(m.max))
	}

	return fmt.Sprintf("Between(%s, %s)", formatValue(m.min), formatValue(m.max))
}
