package match_test

import (
	"testing"
	"time"

	"github.com/grongor/go-muchtest"
	"github.com/grongor/go-muchtest/match"
	"go.uber.org/zap/zapcore"
)

func TestSameSuite(t *testing.T) {
	muchtest.Run(t, new(SameSuite))
}

type SameSuite struct {
	pkgSuite
}

func (s *SameSuite) TestSame() {
	entry1 := &zapcore.Entry{Message: "lol", Time: time.Now()}
	entry2 := entry1

	s.match(match.Same(entry1), entry2)
	s.match(match.Same(*entry1), *entry2)
	s.match(match.Same("Much"), "Much")
	s.match(match.Same(123), 123)
	s.match(match.Same("Much"), "Such", `Same("Much"): not same: string("Such")`)
	s.match(match.Same(123), int16(123), `Same(123): not same: int16(123)`)
}

func (s *SameSuite) TestSamePointer() {
	entry1 := &zapcore.Entry{Message: "lol", Time: time.Date(2022, 8, 26, 13, 14, 58, 0, time.UTC)}
	entry2 := entry1
	str := "Much"
	str1 := &str
	str2 := str1

	s.match(match.SamePointer(entry1), entry2)
	s.match(match.SamePointer(str1), str2)
	s.match(match.SamePointer(*entry1), *entry2, `Same(zapcore.Entry{Message:"lol", Time:2022-08-26 13:14:58}): not `+
		`pointer: zapcore.Entry{Message:"lol", Time:2022-08-26 13:14:58}`)
	s.match(match.SamePointer(entry1), &zapcore.EntryCaller{File: "much.go"},
		`Same(*zapcore.Entry{Message:"lol", Time:2022-08-26 13:14:58}): not same type: *zapcore.EntryCaller(undefined)`)
	s.match(
		match.SamePointer(entry1),
		&zapcore.Entry{Message: "lol", Time: time.Date(2022, 8, 26, 13, 14, 58, 0, time.UTC)},
		`Same(*zapcore.Entry{Message:"lol", Time:2022-08-26 13:14:58}): `+
			`not same: *zapcore.Entry{Message:"lol", Time:2022-08-26 13:14:58}`,
	)
}
