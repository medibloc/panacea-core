package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DistributeRewardToOracle distributes the reward to the oracle and its delegator
func (k Keeper) DistributeRewardToOracle(ctx sdk.Context, oracleAddr string, reward sdk.DecCoins) {
	oracleAccAddr, err := sdk.AccAddressFromBech32(oracleAddr)
	if err != nil {
		panic(err)
	}

	oracleValAddr := sdk.ValAddress(oracleAccAddr.Bytes())
	validator, ok := k.stakingKeeper.GetValidator(ctx, oracleValAddr)
	if !ok {
		panic(fmt.Sprintf("failed to retrieve validator information. address: %s", oracleAddr))
	}

	k.distrKeeper.AllocateTokensToValidator(ctx, validator, reward)
}
