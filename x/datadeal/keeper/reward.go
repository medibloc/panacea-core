package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/medibloc/panacea-core/v2/types/assets"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
)

// DistributeRewards distributes reward to oracles for data verification and delivery
func (k Keeper) DistributeRewards(ctx sdk.Context, dealID uint64, oracles map[string]*oracletypes.OracleValidatorInfo) {
	deal, err := k.GetDeal(ctx, dealID)
	if err != nil {
		panic(err)
	}

	dealAddress, err := sdk.AccAddressFromBech32(deal.GetAddress())
	if err != nil {
		panic(err)
	}

	totalBudget := deal.GetBudget().Amount.ToDec()
	maxNumData := sdk.NewIntFromUint64(deal.GetMaxNumData()).ToDec()
	oracleCommission := k.oracleKeeper.GetParams(ctx).OracleCommissionRate
	pricePerData := sdk.NewCoin(assets.MicroMedDenom, totalBudget.Quo(maxNumData).TruncateDec().Mul(oracleCommission).TruncateInt())

	dealBalance := k.bankKeeper.GetBalance(ctx, dealAddress, assets.MicroMedDenom)
	if dealBalance.IsLT(pricePerData) {
		panic(fmt.Sprintf("deal's balanace is not enough"))
	}

	totalReward := sdk.NewDecCoinsFromCoins(pricePerData)
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, dealAddress, distrtypes.ModuleName, sdk.Coins{pricePerData}); err != nil {
		panic(err)
	}

	// calculate total voting power
	totalVotingPower := sdk.ZeroInt()
	for _, oracle := range oracles {
		if oracle.IsPossibleVote() {
			totalVotingPower = totalVotingPower.Add(oracle.BondedTokens)
		}
	}
	totalVotingPowerDec := totalVotingPower.ToDec()

	for _, oracle := range oracles {
		if oracle.IsPossibleVote() {
			bondedDec := oracle.BondedTokens.ToDec()
			fraction := bondedDec.Quo(totalVotingPowerDec)
			reward := totalReward.MulDec(fraction)
			k.oracleKeeper.DistributeRewardToOracle(ctx, oracle.Address, reward)
		}
	}
}
