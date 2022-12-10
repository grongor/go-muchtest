package muchtest_test

import (
	"regexp"
	"runtime"
	"testing"

	"github.com/grongor/go-muchtest"
	"github.com/grongor/go-muchtest/internal"
	"github.com/grongor/go-muchtest/match"
	"github.com/grongor/go-muchtest/mocks"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestZapSuite(t *testing.T) {
	muchtest.Run(t, new(ZapSuite))
}

type ZapSuite struct {
	muchtest.Suite

	TestingT *mocks.TestingT

	z *muchtest.Zap
}

func (s *ZapSuite) BeforeTest(suiteName, testName string) {
	s.z = muchtest.NewZap(s.TestingT)

	s.TestingT.EXPECT().Helper().Maybe()
}

func (s *ZapSuite) TestGetGetAtAndUse() {
	const getMultiple = "You can't call Get(), GetAt() or Use() multiple times"

	for _, test := range []string{"get", "getAt", "use"} {
		s.Run(test, func() {

			switch test {
			case "get":
				s.S.NotNil(s.z.Get().Sugar())
			case "getAt":
				s.S.NotNil(s.z.GetAt(zap.WarnLevel))
			case "use":
				s.S.NotNil(s.z.Use(nil, nil))
			}

			s.expectErr(getMultiple, func() { s.z.Get().Sugar() })
			s.expectErr(getMultiple, func() { s.z.GetAt(zap.InfoLevel) })
			s.expectErr(getMultiple, func() { s.z.GetAt(zap.WarnLevel) })
			s.expectErr(getMultiple, func() { s.z.Use(nil, nil) })
		})
	}
}

func (s *ZapSuite) TestObservedLogs() {
	s.expectErr("First call Get(), GetAt() or Use() methods", func() { s.z.ObservedLogs() })

	logger := s.z.Get().Sugar()
	s.S.NotNil(logger)

	observedLogs := s.z.ObservedLogs()
	s.S.NotNil(observedLogs)

	logger.Infow("msg", "key", "value")
	s.S.Match(
		match.MapExact(0, match.All(
			match.Map("Entry", match.Map("Message", "msg", "Level", zap.InfoLevel)),
			match.Method("ContextMap", match.MapExact("key", "value")),
		)),
		observedLogs.All(),
	)

	s.expectErr("You can't call ObservedLogs() multiple times", func() { s.z.ObservedLogs() })
	s.expectErr("You can't combine ObservedLogs() and AssertNext*() methods", func() { s.z.AssertNext(zap.InfoLevel, "") })
	s.expectErr(
		"You can't combine ObservedLogs() and AssertNext*() methods",
		func() { s.z.AssertNamed("", zap.InfoLevel, "") },
	)
}

func (s *ZapSuite) TestAssertNext() {
	logger := s.z.Get().Sugar()

	s.expectErr("There are no more logged entries", func() { s.z.AssertNext(zap.InfoLevel, "") })

	logger.Infow("msg")
	s.expectErr(
		`Equal(zapcore.Level(warn)): not equal: zapcore.Level(info)`,
		func() { s.z.AssertNext(zap.WarnLevel, "") },
	)

	logger.Infow("msg")
	s.expectErr(`Equal(""): not equal: "msg"`, func() { s.z.AssertNext(zap.InfoLevel, "") })

	logger.Infow("msg", "key", 123)
	s.expectErr(`Len(0): got 1: map[string]any{"key":123}`, func() { s.z.AssertNext(zap.InfoLevel, "msg") })

	logger.Infow("msg", "key", 123)
	msg := `AssertNext(): ctxKeyAndValues must be pairs; eg.: "duration", 5*time.Second, "status", "ok"`
	s.expectErr(msg, func() { s.z.AssertNext(zap.InfoLevel, "msg", "lorem") })

	logger.Infow("msg", "key", 123)
	msg = `ContainsKeyValue("lorem", "wow"): no such key in: map[string]any{"key":123}`
	s.expectErr(msg, func() { s.z.AssertNext(zap.InfoLevel, "msg", "lorem", "wow") })

	logger.Infow("msg", "key", 123)
	msg = `ContainsKeyValue("key", "wow"): value not matched: 123`
	s.expectErr(msg, func() { s.z.AssertNext(zap.InfoLevel, "msg", "key", "wow") })

	logger.Infow("msg", "key", 123)
	s.z.AssertNext(zap.InfoLevel, "msg", "key", 123)

	logger.Infow("msg", "key", 123)
	s.z.AssertNext(zap.InfoLevel, "msg", match.MapExact("key", 123))
}

func (s *ZapSuite) TestAssertNextNamed_NoName() {
	s.z.Get().Sugar().Infow("msg")
	s.z.AssertNamed("", zap.InfoLevel, "msg")
}

func (s *ZapSuite) TestAssertNextNamed() {
	const name = "much"

	logger := s.z.Get().Sugar().Named(name)

	s.expectErr("There are no more logged entries", func() { s.z.AssertNamed("", zap.WarnLevel, "") })

	logger.Infow("msg")
	s.expectErr(`Equal(""): not equal: "much"`, func() { s.z.AssertNamed("", zap.WarnLevel, "") })

	logger.Infow("msg")
	s.expectErr(
		`Equal(zapcore.Level(warn)): not equal: zapcore.Level(info)`,
		func() { s.z.AssertNamed(name, zap.WarnLevel, "") },
	)

	logger.Infow("msg")
	s.expectErr(`Equal(""): not equal: "msg"`, func() { s.z.AssertNamed(name, zap.InfoLevel, "") })

	logger.Infow("msg", "key", 123)
	s.expectErr(`Len(0): got 1: map[string]any{"key":123}`, func() { s.z.AssertNamed(name, zap.InfoLevel, "msg") })

	logger.Infow("msg", "key", 123)
	msg := `AssertNamed(): ctxKeyAndValues must be pairs; eg.: "duration", 5*time.Second, "status", "ok"`
	s.expectErr(msg, func() { s.z.AssertNamed(name, zap.InfoLevel, "msg", "lorem") })

	logger.Infow("msg", "key", 123)
	msg = `ContainsKeyValue("lorem", "wow"): no such key in: map[string]any{"key":123}`
	s.expectErr(msg, func() { s.z.AssertNamed(name, zap.InfoLevel, "msg", "lorem", "wow") })

	logger.Infow("msg", "key", 123)
	msg = `ContainsKeyValue("key", "wow"): value not matched: 123`
	s.expectErr(msg, func() { s.z.AssertNamed(name, zap.InfoLevel, "msg", "key", "wow") })

	logger.Infow("msg", "key", 123)
	s.z.AssertNamed(name, zap.InfoLevel, "msg", "key", 123)
}

func (s *ZapSuite) TestSkipNext() {
	logger := s.z.Get().Sugar()
	logger.Info("msg")
	s.z.SkipNext()
	s.z.AssertNoNextEntry()

	logger.Info("msg")
	s.z.SkipNext(2)
	logger.Info("msg")
	s.z.AssertNoNextEntry()

	s.expectErr(
		"Parameter to SkipNext() must be a single positive integer, or nothing (equivalent to 1)",
		func() { s.z.SkipNext(0) },
	)
}

func (s *ZapSuite) TestAssertNoNextEntry() {
	logger := s.z.Get().Sugar()
	s.z.AssertNoNextEntry()

	logger.Info("msg")
	s.expectErr(`AssertNoNextEntry(): there are entries available (1)`, func() { s.z.AssertNoNextEntry() })
}

func (s *ZapSuite) TestIgnoreMissingContext() {
	logger := s.z.Get().Sugar()

	s.expectErr(
		"IgnoreMissingContext(): parameter must be a single bool, or nothing (equivalent to true)",
		func() { s.z.IgnoreMissingContext(true, true) },
	)

	logger.Infow("msg", "key", 123)

	s.z.IgnoreMissingContext()
	s.z.AssertNext(zap.InfoLevel, "msg")

	s.z.IgnoreMissingContext(false)
	logger.Infow("msg", "key", 123)

	s.expectErr(`Len(0): got 1: map[string]any{"key":123}`, func() { s.z.AssertNext(zap.InfoLevel, "msg") })

	logger.Infow("msg", "key", 123)
	s.z.IgnoreMissingContext(true)
	s.z.AssertNext(zap.InfoLevel, "msg")
}

func (s *ZapSuite) TestOrderless() {
	logger := s.z.Get().Sugar()

	s.expectErr(
		"Orderless(): parameter must be a single bool, or nothing (equivalent to true)",
		func() { s.z.Orderless(true, true) },
	)

	logger.Info("msg")
	logger.Error("err")

	s.z.Orderless()
	s.expectErr("Orderless(): already set to orderless", func() { s.z.Orderless() })
	s.z.AssertNext(zap.ErrorLevel, "err")
	s.z.AssertNext(zap.InfoLevel, "msg")

	s.z.Orderless(false)
	logger.Warn("warn")
	logger.Info("info")

	s.expectErr("Orderless(): already set to ordered", func() { s.z.Orderless(false) })
	s.expectErr(`Equal(zapcore.Level(info)): not equal: zapcore.Level(warn)`, func() { s.z.AssertNext(zap.InfoLevel, "info") })
	s.z.SkipNext()

	logger.Info("msg2")
	s.z.Orderless(true)
	logger.Error("err2")

	s.z.AssertNext(zap.ErrorLevel, "err2")
	s.z.AssertNext(zap.InfoLevel, "msg2")
	s.expectErr(`No entries matched and there are no more logged entries`, func() { s.z.AssertNext(zap.InfoLevel, "") })
}

func (s *ZapSuite) expectErr(desc any, fn func()) {
	s.T().Helper()

	s.TestingT.EXPECT().Errorf(mock.MatchedBy(func(fmt string) bool {
		switch fmt {
		case "\n%s": //,"muchtest: zap: %s"
			return true
		default:
			return false
		}
	}), mock.MatchedBy(func(actual string) bool {
		switch x := desc.(type) {
		case string:
			return regexp.MustCompile(`(?s)\s*Error Trace:(.*?)Error:\s+muchtest: zap: ` + regexp.QuoteMeta(x)).
				MatchString(internal.TrimDiff(actual))
		case *regexp.Regexp:
			panic("wow")
			//return x.MatchString(internal.TrimDiff(actual))
		default:
			return false
		}
	})).Once()

	s.TestingT.EXPECT().FailNow().Once().Run(func(mock.Arguments) { runtime.Goexit() })

	done := make(chan bool)

	go func() {
		defer close(done)

		fn()

		done <- true
	}()

	if notExited := <-done; notExited {
		s.Fail("expected goroutine to exit")
	}
}
