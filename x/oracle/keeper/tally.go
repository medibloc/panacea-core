package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
)

// Tally calculates the result by aggregating the votes taken from the iterator.
// 'voteType' is an unmarshal type of 'voteIterator'.
func (k Keeper) Tally(ctx sdk.Context, voteIterator sdk.Iterator, voteType types.Vote) (*types.TallyResult, error) {
	// If the Iterator is empty, it returns with a default value.
	if !voteIterator.Valid() {
		return types.NewTallyResult(), nil
	}

	tally := types.NewTally()
	k.setOracleValidatorInfosInTally(ctx, tally)

	for ; voteIterator.Valid(); voteIterator.Next() {
		bz := voteIterator.Value()
		err := k.cdc.UnmarshalLengthPrefixed(bz, voteType)
		if err != nil {
			return nil, err
		}

		err = tally.Add(voteType)
		if err != nil {
			return nil, err
		}
	}

	tallyResult := k.calculateTallyResult(ctx, tally)

	return tallyResult, nil
}

func (k Keeper) setOracleValidatorInfosInTally(ctx sdk.Context, tally *types.Tally) {
	oracleValidatorInfos := make(map[string]*types.OracleValidatorInfo)
	k.IterateOracleValidator(ctx, func(info *types.OracleValidatorInfo) bool {
		oracleValidatorInfos[info.Address] = info
		return false
	})
	tally.OracleValidatorInfos = oracleValidatorInfos
}

func (k Keeper) calculateTallyResult(ctx sdk.Context, tally *types.Tally) *types.TallyResult {
	quorum := k.GetParams(ctx).VoteParams.Quorum

	return tally.CalculateTallyResult(quorum)
}
