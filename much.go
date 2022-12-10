package muchtest

import (
	"github.com/stretchr/testify/require"
)

type TestingT interface {
	require.TestingT
	Helper()
}
