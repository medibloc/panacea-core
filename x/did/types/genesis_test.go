package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultGenesisState(t *testing.T) {
	defaultState := DefaultGenesisState()
	require.Empty(t, defaultState.Documents)
}
