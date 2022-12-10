package match_test

import (
	"testing"

	"github.com/grongor/go-muchtest"
)

func TestMethodSuite(t *testing.T) {
	muchtest.Run(t, new(MethodSuite))
}

type MethodSuite struct {
	pkgSuite
}
