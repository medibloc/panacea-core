package types_test

import (
	"fmt"
	"github.com/medibloc/panacea-core/v2/x/market/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewDealAddress(t *testing.T) {
	var firstDealId uint64 = 1
	var secondDealId uint64 = 1

	firstDealAddress := types.NewDealAddress(firstDealId)
	secondDealAddress := types.NewDealAddress(secondDealId)

	fmt.Println(firstDealAddress)
	fmt.Println(secondDealAddress)

	require.Equal(t, firstDealAddress, secondDealAddress)
}