package match_test

import (
	"testing"

	"github.com/grongor/go-muchtest"
	"github.com/grongor/go-muchtest/match"
)

func TestBetweenSuite(t *testing.T) {
	muchtest.Run(t, new(BetweenSuite))
}

type BetweenSuite struct {
	pkgSuite
}

func (s *BetweenSuite) TestBetween_Loose() {
	s.match(match.Between(6, 9), 1+2i, "Between(6, 9): not comparable: complex128((1+2i))")

	// int
	s.match(match.Between(6, int8(9)), int16(-12), "Between(6, 9): outside range: -12")
	s.match(match.Between(6, int8(9)), int16(6))
	s.match(match.Between(6, 9), 7)
	s.match(match.Between(int8(6), 9), int16(9))
	s.match(match.Between(6, 9), 34, "Between(6, 9): outside range: 34")

	// uint
	s.match(match.Between(uint(6), uint8(9)), uint16(0), "Between(6, 9): outside range: 0")
	s.match(match.Between(uint(6), uint8(9)), uint16(6))
	s.match(match.Between(uint(6), uint(9)), uint(7))
	s.match(match.Between(uint8(6), uint(9)), uint16(9))
	s.match(match.Between(uint(6), uint(9)), uint(34), "Between(6, 9): outside range: 34")

	// float
	s.match(match.Between(12.3, float32(13.4)), float64(-1.23e+10), "Between(12.3, 13.4): outside range: -1.23e+10")
	s.match(match.Between(12.3, float32(13.4)), float64(12.3))
	s.match(match.Between(float32(12.3), 13.4), float64(12.8))
	s.match(match.Between(float32(12.3), 13.4), float64(13.4))
	s.match(match.Between(12.3, 13.4), 1.23e+10, "Between(12.3, 13.4): outside range: 1.23e+10")

	// mixed types
	s.match(match.Between(uint(6), 9), 7)
	s.match(match.Between(6, uint(9)), 7)
	s.match(match.Between(6, 9), uint(7))
	s.match(match.Between(6, 9), 7)
	s.match(match.Between(6.3, 9), 7)
	s.match(match.Between(6, 9.4), 7)
	s.match(match.Between(6, 9), 7.5)
	s.match(match.Between(uint(6), 9), 4, "Between(6, 9): outside range: 4")
	s.match(match.Between(6, uint(9)), 4, "Between(6, 9): outside range: 4")
	s.match(match.Between(6, 9), uint(4), "Between(6, 9): outside range: 4")
	s.match(match.Between(6, 9), 4, "Between(6, 9): outside range: 4")
	s.match(match.Between(6.3, 9), 4, "Between(6.3, 9): outside range: 4")
	s.match(match.Between(6, 9.4), 4, "Between(6, 9.4): outside range: 4")
	s.match(match.Between(6, 9), 4.5, "Between(6, 9): outside range: 4.5")
}

func (s *BetweenSuite) TestBetween_Strict() {
	s.match(match.BetweenStrict(1+2i, 2+3i), 3+4i, "BetweenStrict((1+2i), (2+3i)): not comparable: complex128((3+4i))")

	// int
	s.match(match.BetweenStrict(6, 9), -12, "BetweenStrict(6, 9): outside range: -12")
	s.match(match.BetweenStrict(6, 9), 6)
	s.match(match.BetweenStrict(6, 9), 7)
	s.match(match.BetweenStrict(6, 9), 9)
	s.match(match.BetweenStrict(6, 9), 34, "BetweenStrict(6, 9): outside range: 34")

	// uint
	s.match(match.BetweenStrict(uint(6), uint(9)), uint(0), "BetweenStrict(6, 9): outside range: 0")
	s.match(match.BetweenStrict(uint(6), uint(9)), uint(6))
	s.match(match.BetweenStrict(uint(6), uint(9)), uint(7))
	s.match(match.BetweenStrict(uint(6), uint(9)), uint(9))
	s.match(match.BetweenStrict(uint(6), uint(9)), uint(34), "BetweenStrict(6, 9): outside range: 34")

	// float
	s.match(match.BetweenStrict(12.3, 13.4), -1.23e+10, "BetweenStrict(12.3, 13.4): outside range: -1.23e+10")
	s.match(match.BetweenStrict(12.3, 13.4), 12.3)
	s.match(match.BetweenStrict(12.3, 13.4), 12.8)
	s.match(match.BetweenStrict(12.3, 13.4), 13.4)
	s.match(match.BetweenStrict(12.3, 13.4), 1.23e+10, "BetweenStrict(12.3, 13.4): outside range: 1.23e+10")

	// mixed types
	s.match(match.BetweenStrict(int8(6), 9), 6,
		"BetweenStrict(6, 9): min and max have different types: min(int8) max(int)")
	s.match(match.BetweenStrict(6, uint(9)), 6,
		"BetweenStrict(6, 9): min and max have different types: min(int) max(uint)")
	s.match(match.BetweenStrict(float32(6.1), 9), 6,
		"BetweenStrict(6.1, 9): min and max have different types: min(float32) max(int)")
	s.match(match.BetweenStrict(6, 9), int8(7), "BetweenStrict(6, 9): different type: int8(7)")
	s.match(match.BetweenStrict(6, 9), uint(7), "BetweenStrict(6, 9): different type: uint(7)")
	s.match(match.BetweenStrict(6, 9), float32(7.3), "BetweenStrict(6, 9): different type: float32(7.3)")
}
