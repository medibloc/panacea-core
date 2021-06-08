package types_test

import (
	"testing"

	"github.com/medibloc/panacea-core/x/did/types"
	"github.com/stretchr/testify/require"
)

func TestDefaultGenesisState(t *testing.T) {
	defaultState := types.DefaultGenesis()
	require.Empty(t, defaultState.Documents)
}
