package keeper

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/medibloc/panacea-core/v2/types/assets"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
)

func (k Keeper) DistributeVerificationRewards(ctx sdk.Context, dataSale *types.DataSale) error {
	deal, err := k.GetDeal(ctx, dataSale.DealId)
	if err != nil {
		return err
	}

	totalBudget := deal.GetBudget().Amount.ToDec()
	maxNumData := sdk.NewIntFromUint64(deal.GetMaxNumData()).ToDec()
	pricePerData := totalBudget.Quo(maxNumData).TruncateDec()

	oracleCommissionRate := k.oracleKeeper.GetParams(ctx).OracleCommissionRate
	sellerReward := sdk.NewCoin(assets.MicroMedDenom, pricePerData.Mul(sdk.OneDec().Sub(oracleCommissionRate)).TruncateInt())

	if err := k.sendSellerReward(ctx, deal.GetAddress(), dataSale.SellerAddress, sellerReward); err != nil {
		return err
	}

	// TODO distribute rewards to oracles

	return nil
}

func (k Keeper) sendSellerReward(ctx sdk.Context, dealAddress, sellerAddress string, reward sdk.Coin) error {
	dealAccAddr, err := sdk.AccAddressFromBech32(dealAddress)
	if err != nil {
		return err
	}

	sellerAccAddr, err := sdk.AccAddressFromBech32(sellerAddress)
	if err != nil {
		return err
	}

	dealBalance := k.bankKeeper.GetBalance(ctx, dealAccAddr, assets.MicroMedDenom)
	if dealBalance.IsLT(reward) {
		return fmt.Errorf("not enough balance in deal")
	}

	if err := k.bankKeeper.SendCoins(ctx, dealAccAddr, sellerAccAddr, sdk.NewCoins(reward)); err != nil {
		return err
	}

	return nil
}

// DistributeOracleRewards distributes reward to oracles for data verification and delivery
func (k Keeper) DistributeOracleRewards(ctx sdk.Context, dealID uint64, oracles map[string]*oracletypes.OracleValidatorInfo) {
	deal, err := k.GetDeal(ctx, dealID)
	if err != nil {
		panic(err)
	}

	dealAddress, err := sdk.AccAddressFromBech32(deal.GetAddress())
	if err != nil {
		panic(err)
	}

	// calculate oracle commission
	totalBudget := deal.GetBudget().Amount.ToDec()
	maxNumData := sdk.NewIntFromUint64(deal.GetMaxNumData()).ToDec()
	oracleCommission := k.oracleKeeper.GetParams(ctx).OracleCommissionRate
	pricePerData := sdk.NewCoin(assets.MicroMedDenom, totalBudget.Quo(maxNumData).TruncateDec().Mul(oracleCommission).TruncateInt())

	dealBalance := k.bankKeeper.GetBalance(ctx, dealAddress, assets.MicroMedDenom)
	if dealBalance.IsLT(pricePerData) {
		panic(fmt.Errorf("deal's balanace is not enough"))
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

	// distribute rewards
	for _, oracle := range oracles {
		if oracle.IsPossibleVote() {
			bondedDec := oracle.BondedTokens.ToDec()
			fraction := bondedDec.Quo(totalVotingPowerDec)
			reward := totalReward.MulDec(fraction)
			k.oracleKeeper.DistributeRewardToOracle(ctx, oracle.Address, reward)
		}
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			oracletypes.EventTypeOracleReward,
			sdk.NewAttribute(types.AttributeKeyDealID, strconv.FormatUint(dealID, 10)),
		),
	)
}
