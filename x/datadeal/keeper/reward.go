package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/medibloc/panacea-core/v2/types/assets"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
)

func (k Keeper) DistributeVerificationRewards(ctx sdk.Context, dataSale *types.DataSale, voters []*oracletypes.VoterInfo) error {
	deal, err := k.GetDeal(ctx, dataSale.DealId)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrDistrVerificationRewards, err.Error())
	}

	dealAccAddr, err := sdk.AccAddressFromBech32(deal.GetAddress())
	if err != nil {
		return sdkerrors.Wrapf(types.ErrDistrVerificationRewards, err.Error())
	}

	sellerAccAddr, err := sdk.AccAddressFromBech32(dataSale.GetSellerAddress())
	if err != nil {
		return sdkerrors.Wrapf(types.ErrDistrVerificationRewards, err.Error())
	}

	totalBudget := deal.GetBudget().Amount.ToDec()
	maxNumData := sdk.NewIntFromUint64(deal.GetMaxNumData()).ToDec()
	pricePerData := totalBudget.Quo(maxNumData).TruncateDec()

	dealBalance := k.bankKeeper.GetBalance(ctx, dealAccAddr, assets.MicroMedDenom)
	if dealBalance.IsLT(sdk.NewCoin(assets.MicroMedDenom, pricePerData.TruncateInt())) {
		return sdkerrors.Wrapf(types.ErrDistrVerificationRewards, "not enough balance in deal")
	}

	oracleCommissionRate := k.oracleKeeper.GetParams(ctx).OracleCommissionRate
	sellerReward := sdk.NewCoin(assets.MicroMedDenom, pricePerData.Mul(sdk.OneDec().Sub(oracleCommissionRate)).TruncateInt())
	// 50% of oracle commission for data verification
	// remain is for data delivery
	oracleRewards := sdk.NewCoin(assets.MicroMedDenom, pricePerData.Mul(oracleCommissionRate).Mul(sdk.NewDecWithPrec(5, 1)).TruncateInt())

	// send to seller
	if err := k.bankKeeper.SendCoins(ctx, dealAccAddr, sellerAccAddr, sdk.NewCoins(sellerReward)); err != nil {
		return sdkerrors.Wrapf(types.ErrDistrVerificationRewards, err.Error())
	}

	// send to oracles
	if err := k.distributeOracleRewards(ctx, dealAccAddr, voters, oracleRewards); err != nil {
		return sdkerrors.Wrapf(types.ErrDistrVerificationRewards, err.Error())
	}

	return nil
}

// distributeOracleRewards distributes reward to oracles for data verification and delivery
func (k Keeper) distributeOracleRewards(ctx sdk.Context, dealAccAddr sdk.AccAddress, oracles []*oracletypes.VoterInfo, rewards sdk.Coin) error {
	// send reward to distribution module
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, dealAccAddr, distrtypes.ModuleName, sdk.Coins{rewards}); err != nil {
		return err
	}

	// calculate total voting power
	totalVotingPower := sdk.ZeroInt()
	for _, oracle := range oracles {
		totalVotingPower = totalVotingPower.Add(oracle.VotingPower)
	}
	totalVotingPowerDec := totalVotingPower.ToDec()

	// distribute rewards proportional to its voting power
	totalRewards := sdk.NewDecCoinsFromCoins(rewards)
	for _, oracle := range oracles {
		bondedDec := oracle.VotingPower.ToDec()
		reward := totalRewards.MulDec(bondedDec).QuoDecTruncate(totalVotingPowerDec)
		k.oracleKeeper.DistributeRewardToOracle(ctx, oracle.GetVoterAddress(), reward)
	}

	return nil
}
