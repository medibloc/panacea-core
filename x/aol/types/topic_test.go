package types_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/x/aol/types"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("panacea", "panaceapub")
	config.Seal()

	os.Exit(m.Run())
}

func TestIncrease_Decrease(t *testing.T) {
	topic := types.Topic{
		TotalRecords: 0,
		TotalWriters: 0,
		Description: "test topic",
	}

	topic = topic.IncreaseTotalWriters()
	require.Equal(t, uint64(1), topic.TotalWriters)
	require.Equal(t, uint64(0), topic.TotalRecords)
	require.Equal(t, "test topic", topic.GetDescription())

	topic = topic.IncreaseTotalRecords()
	require.Equal(t, uint64(1), topic.TotalWriters)
	require.Equal(t, uint64(1), topic.TotalRecords)
	require.Equal(t, "test topic", topic.GetDescription())

	topic = topic.DecreaseTotalWriters()
	require.Equal(t, uint64(0), topic.TotalWriters)
	require.Equal(t, uint64(1), topic.TotalRecords)
	require.Equal(t, "test topic", topic.GetDescription())
}