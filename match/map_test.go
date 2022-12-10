package match_test

import (
	"testing"

	"github.com/grongor/go-muchtest"
)

func TestMapSuite(t *testing.T) {
	muchtest.Run(t, new(MapSuite))
}

type MapSuite struct {
	pkgSuite
}
